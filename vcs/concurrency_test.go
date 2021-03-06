package vcs

import (
	"runtime"
	"testing"
	"time"
)

// This test checks that you have compiled libgit2 with -DTHREADSAFE=ON. If you
// didn't, then this test will crash in cgo.
//
// Example stack trace: https://gist.github.com/sqs/ce913ce35599d3377c11
func TestRepository_LibGit2_Concurrency(t *testing.T) {
	p := runtime.GOMAXPROCS(0)
	if p == 1 {
		t.Skip("no point in testing concurrency with GOMAXPROCS=1")
	}

	origRepo := makeGitRepositoryLibGit2(t,
		"echo hello > a",
		"git add a",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m foo --author='a <a@a.com>' --date 2006-01-02T15:04:05Z",
		"GIT_COMMITTER_NAME=a GIT_COMMITTER_EMAIL=a@a.com GIT_COMMITTER_DATE=2006-01-02T15:04:05Z git commit --allow-empty -m bar --author='a <a@a.com>' --date 2006-01-02T17:04:05Z",
	)

	n := 10
	start := time.Now()
	duration := 5 * time.Second

	for i := 0; i < n; i++ {
		go func() {
			for {
				if time.Since(start) > duration {
					return
				}
				repo, err := OpenGitRepositoryLibGit2(origRepo.dir)
				if err != nil {
					t.Error(err)
					return
				}

				commitID, err := repo.ResolveRevision("master")
				if err != nil {
					t.Error(err)
					return
				}

				_, err = repo.GetCommit(commitID)
				if err != nil {
					t.Error(err)
					return
				}

				_, err = repo.CommitLog(commitID)
				if err != nil {
					t.Error(err)
					return
				}

				fs, err := repo.FileSystem(commitID)
				if err != nil {
					t.Error(err)
					return
				}

				entries, err := fs.ReadDir(".")
				if err != nil {
					t.Error(err)
					return
				}

				if len(entries) != 1 {
					t.Errorf("got entries %v, want 1 entry", entries)
				}

				for _, entry := range entries {
					f, err := fs.Open(entry.Name())
					if err != nil {
						t.Error(err)
						return
					}
					f.Close()
				}
			}
		}()
	}

	time.Sleep(duration + 500*time.Millisecond)
}
