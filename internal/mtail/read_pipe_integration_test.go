// Copyright 2019 Google Inc. All Rights Reserved.
// This file is available under the Apache license.

package mtail_test

import (
	"fmt"
	"os"
	"path"
	"syscall"
	"testing"

	"github.com/golang/glog"
	"github.com/google/mtail/internal/mtail"
	"github.com/google/mtail/internal/testutil"
	"golang.org/x/sys/unix"
)

func TestReadFromPipe(t *testing.T) {
	testutil.SkipIfShort(t)
	for _, test := range mtail.LogWatcherTestTable {
		t.Run(fmt.Sprintf("%s %v", test.PollInterval, test.EnableFsNotify), func(t *testing.T) {
			tmpDir, rmTmpDir := testutil.TestTempDir(t)
			defer rmTmpDir()

			logDir := path.Join(tmpDir, "logs")
			progDir := path.Join(tmpDir, "progs")
			testutil.FatalIfErr(t, os.Mkdir(logDir, 0700))
			testutil.FatalIfErr(t, os.Mkdir(progDir, 0700))
			defer testutil.TestChdir(t, logDir)()

			logFile := path.Join(logDir, "logpipe")

			testutil.FatalIfErr(t, unix.Mkfifo(logFile, 0600))

			f, err := os.OpenFile(logFile, os.O_RDWR|syscall.O_NONBLOCK, 0600)
			testutil.FatalIfErr(t, err)
			defer func() {
				testutil.FatalIfErr(t, f.Close())
			}()

			m, stopM := mtail.TestStartServer(t, test.PollInterval, test.EnableFsNotify, mtail.LogPathPatterns(logDir+"/*"), mtail.ProgramPath(progDir))
			defer stopM()

			lineCountCheck := m.ExpectMetricDeltaWithDeadline("lines_total", 3)

			n, err := f.WriteString("1\n2\n3\n")
			testutil.FatalIfErr(t, err)
			glog.Infof("Wrote %d bytes", n)

			lineCountCheck()
		})
	}
}
