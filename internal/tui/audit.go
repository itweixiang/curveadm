/*
 *  Copyright (c) 2021 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: CurveAdm
 * Created Date: 2022-05-23
 * Author: Jingli Chen (Wine93)
 */

package tui

import (
	"strconv"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/internal/storage"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	RESULT_ABORT   = "ABORT"
	RESULT_SUCCESS = "SUCCESS"
	RESULT_FAIL    = "FAIL"
)

func resultDecorate(message string) string {
	if message == RESULT_SUCCESS {
		return color.GreenString(message)
	} else if message == RESULT_FAIL {
		return color.RedString(message)
	}
	return message
}

func FormatAuditLogs(auditLogs []storage.AuditLog) string {
	lines := [][]interface{}{}
	title := []string{"Id", "Result", "Execute Time", "Command"}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	for i := 0; i < len(auditLogs); i++ {
		line := []interface{}{}
		auditLog := auditLogs[i]
		// id
		line = append(line, strconv.Itoa(auditLog.Id))
		// result
		result := RESULT_ABORT
		if auditLog.Status == 0 {
			result = RESULT_SUCCESS
		} else if auditLog.Status == 1 {
			result = RESULT_FAIL
		}
		line = append(line, tuicommon.DecorateMessage{Message: result, Decorate: resultDecorate})
		// execute time
		line = append(line, auditLog.ExecuteTime.Format("2006-01-02 15:04:05"))
		// command
		line = append(line, auditLog.Command)

		lines = append(lines, line)
	}

	output := tuicommon.FixedFormat(lines, 2)
	return output
}
