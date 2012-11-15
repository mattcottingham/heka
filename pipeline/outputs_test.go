/***** BEGIN LICENSE BLOCK *****
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this file,
# You can obtain one at http://mozilla.org/MPL/2.0/.
#
# The Initial Developer of the Original Code is the Mozilla Foundation.
# Portions created by the Initial Developer are Copyright (C) 2012
# the Initial Developer. All Rights Reserved.
#
# Contributor(s):
#   Rob Miller (rmiller@mozilla.com)
#
# ***** END LICENSE BLOCK *****/
package pipeline

import (
	"code.google.com/p/gomock/gomock"
	"encoding/json"
	"fmt"
	"github.com/bitly/go-notify"
	gs "github.com/rafrombrc/gospec/src/gospec"
	"github.com/rafrombrc/gospec/src/gospec"
	"io/ioutil"
	"os"
	"runtime"
	"time"
)

func OutputsSpec(c gospec.Context) {
	t := new(SimpleT)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	c.Specify("A FileOutput", func() {
		fileOutput := new(FileOutput)
		tmpFileName := fmt.Sprintf("fileoutput-test-%d", time.Now().UnixNano())
		tmpFilePath := fmt.Sprint(os.TempDir(), string(os.PathSeparator),
			tmpFileName)
		config := fileOutput.ConfigStruct().(*FileOutputConfig)
		config.Path = tmpFilePath

		msg := getTestMessage()
		pipelinePack := getTestPipelinePack()
		pipelinePack.Message = msg
		pipelinePack.Decoded = true

		closeAndStop := func(tmpFile *os.File) {
			tmpFile.Close()
			notify.Post(STOP, nil)
		}

		// The tests are littered w/ scheduler yields (i.e. runtime.Gosched()
		// calls) so we give the output a chance to respond to the messages
		// we're sending.

		c.Specify("writes text", func() {
			err := fileOutput.Init(config)
			c.Assume(err, gs.IsNil)
			runtime.Gosched()

			c.Specify("by default", func() {
				fileOutput.Deliver(pipelinePack)
				runtime.Gosched()
				tmpFile, err := os.Open(tmpFilePath)
				defer closeAndStop(tmpFile)
				c.Assume(err, gs.IsNil)
				contents, err := ioutil.ReadAll(tmpFile)
				c.Assume(err, gs.IsNil)
				c.Expect(string(contents), gs.Equals, msg.Payload+"\n")
			})

			c.Specify("w/ a prepended timestamp when specified", func() {
				fileOutput.prefix_ts = true
				fileOutput.Deliver(pipelinePack)
				todayStr := time.Now().Format("[2006/Jan/02:")
				runtime.Gosched()
				tmpFile, err := os.Open(tmpFilePath)
				defer closeAndStop(tmpFile)
				c.Expect(err, gs.IsNil)
				contents, err := ioutil.ReadAll(tmpFile)
				strContents := string(contents)
				c.Expect(strContents, StringContains, msg.Payload)
				// NOTE: This will fail if you hit the very unusual edge case
				// of the date changing btn when the message is written to the
				// file and the calculation of `todayStr`, immediately
				// afterwards.
				c.Expect(strContents, StringStartsWith, todayStr)
			})
		})

		c.Specify("writes JSON", func() {
			config.Format = "json"
			err := fileOutput.Init(config)
			c.Assume(err, gs.IsNil)
			runtime.Gosched()

			c.Specify("when specified", func() {
				fileOutput.Deliver(pipelinePack)
				runtime.Gosched()
				tmpFile, err := os.Open(tmpFilePath)
				defer closeAndStop(tmpFile)
				c.Assume(err, gs.IsNil)
				contents, err := ioutil.ReadAll(tmpFile)
				c.Assume(err, gs.IsNil)
				msgJson, err := json.Marshal(pipelinePack.Message)
				c.Assume(err, gs.IsNil)
				c.Expect(string(contents), gs.Equals, string(msgJson)+"\n")
			})
		})

		os.Remove(tmpFilePath) // clean up after ourselves
	})
}
