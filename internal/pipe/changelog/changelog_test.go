package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goreleaser/goreleaser/internal/client"
	"github.com/goreleaser/goreleaser/internal/testlib"
	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/goreleaser/pkg/context"
)

func TestDescription(t *testing.T) {
	require.NotEmpty(t, Pipe{}.String())
}

func TestChangelogProvidedViaFlag(t *testing.T) {
	ctx := context.New(config.Project{})
	ctx.ReleaseNotesFile = "testdata/changes.md"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, "c0ff33 coffeee\n", ctx.ReleaseNotes)
}

func TestTemplatedChangelogProvidedViaFlag(t *testing.T) {
	ctx := context.New(config.Project{})
	ctx.ReleaseNotesFile = "testdata/changes.md"
	ctx.ReleaseNotesTmpl = "testdata/changes-templated.md"
	ctx.Git.CurrentTag = "v0.0.1"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Equal(t, "c0ff33 coffeee v0.0.1\n", ctx.ReleaseNotes)
}

func TestChangelogProvidedViaFlagDoesntExist(t *testing.T) {
	ctx := context.New(config.Project{})
	ctx.ReleaseNotesFile = "testdata/changes.nope"
	require.EqualError(t, Pipe{}.Run(ctx), "open testdata/changes.nope: no such file or directory")
}

func TestReleaseHeaderProvidedViaFlagDoesntExist(t *testing.T) {
	ctx := context.New(config.Project{})
	ctx.ReleaseHeaderFile = "testdata/header.nope"
	require.EqualError(t, Pipe{}.Run(ctx), "open testdata/header.nope: no such file or directory")
}

func TestReleaseFooterProvidedViaFlagDoesntExist(t *testing.T) {
	ctx := context.New(config.Project{})
	ctx.ReleaseFooterFile = "testdata/footer.nope"
	require.EqualError(t, Pipe{}.Run(ctx), "open testdata/footer.nope: no such file or directory")
}

func TestChangelog(t *testing.T) {
	folder := testlib.Mktmp(t)
	testlib.GitInit(t)
	testlib.GitCommit(t, "first")
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCommit(t, "added feature 1")
	testlib.GitCommit(t, "fixed bug 2")
	testlib.GitCommit(t, "ignored: whatever")
	testlib.GitCommit(t, "docs: whatever")
	testlib.GitCommit(t, "something about cArs we dont need")
	testlib.GitCommit(t, "feat: added that thing")
	testlib.GitCommit(t, "Merge pull request #999 from goreleaser/some-branch")
	testlib.GitCommit(t, "this is not a Merge pull request")
	testlib.GitTag(t, "v0.0.2")
	ctx := context.New(config.Project{
		Dist: folder,
		Changelog: config.Changelog{
			Filters: config.Filters{
				Exclude: []string{
					"docs:",
					"ignored:",
					"(?i)cars",
					"^Merge pull request",
				},
			},
		},
	})
	ctx.Git.PreviousTag = "v0.0.1"
	ctx.Git.CurrentTag = "v0.0.2"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	require.NotContains(t, ctx.ReleaseNotes, "first")
	require.Contains(t, ctx.ReleaseNotes, "added feature 1")
	require.Contains(t, ctx.ReleaseNotes, "fixed bug 2")
	require.NotContains(t, ctx.ReleaseNotes, "docs")
	require.NotContains(t, ctx.ReleaseNotes, "ignored")
	require.NotContains(t, ctx.ReleaseNotes, "cArs")
	require.NotContains(t, ctx.ReleaseNotes, "from goreleaser/some-branch")

	for _, line := range strings.Split(ctx.ReleaseNotes, "\n")[1:] {
		if line == "" {
			continue
		}
		require.Truef(t, strings.HasPrefix(line, "* "), "%q: changelog commit must be a list item", line)
	}

	bts, err := os.ReadFile(filepath.Join(folder, "CHANGELOG.md"))
	require.NoError(t, err)
	require.NotEmpty(t, string(bts))
}

func TestChangelogForGitlab(t *testing.T) {
	folder := testlib.Mktmp(t)
	testlib.GitInit(t)
	testlib.GitCommit(t, "first")
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCommit(t, "added feature 1")
	testlib.GitCommit(t, "fixed bug 2")
	testlib.GitCommit(t, "ignored: whatever")
	testlib.GitCommit(t, "docs: whatever")
	testlib.GitCommit(t, "something about cArs we dont need")
	testlib.GitCommit(t, "feat: added that thing")
	testlib.GitCommit(t, "Merge pull request #999 from goreleaser/some-branch")
	testlib.GitCommit(t, "this is not a Merge pull request")
	testlib.GitTag(t, "v0.0.2")
	ctx := context.New(config.Project{
		Dist: folder,
		Changelog: config.Changelog{
			Filters: config.Filters{
				Exclude: []string{
					"docs:",
					"ignored:",
					"(?i)cars",
					"^Merge pull request",
				},
			},
		},
	})
	ctx.TokenType = context.TokenTypeGitLab
	ctx.Git.PreviousTag = "v0.0.1"
	ctx.Git.CurrentTag = "v0.0.2"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	require.NotContains(t, ctx.ReleaseNotes, "first")
	require.Contains(t, ctx.ReleaseNotes, "added feature 1") // no whitespace because its the last entry of the changelog
	require.Contains(t, ctx.ReleaseNotes, "fixed bug 2   ")  // whitespaces are on purpose
	require.NotContains(t, ctx.ReleaseNotes, "docs")
	require.NotContains(t, ctx.ReleaseNotes, "ignored")
	require.NotContains(t, ctx.ReleaseNotes, "cArs")
	require.NotContains(t, ctx.ReleaseNotes, "from goreleaser/some-branch")

	bts, err := os.ReadFile(filepath.Join(folder, "CHANGELOG.md"))
	require.NoError(t, err)
	require.NotEmpty(t, string(bts))
}

func TestChangelogSort(t *testing.T) {
	testlib.Mktmp(t)
	testlib.GitInit(t)
	testlib.GitCommit(t, "whatever")
	testlib.GitTag(t, "v0.9.9")
	testlib.GitCommit(t, "c: commit")
	testlib.GitCommit(t, "a: commit")
	testlib.GitCommit(t, "b: commit")
	testlib.GitTag(t, "v1.0.0")
	ctx := context.New(config.Project{
		Changelog: config.Changelog{},
	})
	ctx.Git.PreviousTag = "v0.9.9"
	ctx.Git.CurrentTag = "v1.0.0"

	for _, cfg := range []struct {
		Sort    string
		Entries []string
	}{
		{
			Sort: "",
			Entries: []string{
				"b: commit",
				"a: commit",
				"c: commit",
			},
		},
		{
			Sort: "asc",
			Entries: []string{
				"a: commit",
				"b: commit",
				"c: commit",
			},
		},
		{
			Sort: "desc",
			Entries: []string{
				"c: commit",
				"b: commit",
				"a: commit",
			},
		},
	} {
		t.Run("changelog sort='"+cfg.Sort+"'", func(t *testing.T) {
			ctx.Config.Changelog.Sort = cfg.Sort
			entries, err := buildChangelog(ctx)
			require.NoError(t, err)
			require.Len(t, entries, len(cfg.Entries))
			var changes []string
			for _, line := range entries {
				changes = append(changes, extractCommitInfo(line))
			}
			require.EqualValues(t, cfg.Entries, changes)
		})
	}
}

func TestChangelogInvalidSort(t *testing.T) {
	ctx := context.New(config.Project{
		Changelog: config.Changelog{
			Sort: "dope",
		},
	})
	require.EqualError(t, Pipe{}.Run(ctx), ErrInvalidSortDirection.Error())
}

func TestChangelogOfFirstRelease(t *testing.T) {
	testlib.Mktmp(t)
	testlib.GitInit(t)
	msgs := []string{
		"initial commit",
		"another one",
		"one more",
		"and finally this one",
	}
	for _, msg := range msgs {
		testlib.GitCommit(t, msg)
	}
	testlib.GitTag(t, "v0.0.1")
	ctx := context.New(config.Project{})
	ctx.Git.CurrentTag = "v0.0.1"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	for _, msg := range msgs {
		require.Contains(t, ctx.ReleaseNotes, msg)
	}
}

func TestChangelogFilterInvalidRegex(t *testing.T) {
	testlib.Mktmp(t)
	testlib.GitInit(t)
	testlib.GitCommit(t, "commitssss")
	testlib.GitTag(t, "v0.0.3")
	testlib.GitCommit(t, "commitzzz")
	testlib.GitTag(t, "v0.0.4")
	ctx := context.New(config.Project{
		Changelog: config.Changelog{
			Filters: config.Filters{
				Exclude: []string{
					"(?iasdr4qasd)not a valid regex i guess",
				},
			},
		},
	})
	ctx.Git.PreviousTag = "v0.0.3"
	ctx.Git.CurrentTag = "v0.0.4"
	require.EqualError(t, Pipe{}.Run(ctx), "error parsing regexp: invalid or unsupported Perl syntax: `(?ia`")
}

func TestChangelogNoTags(t *testing.T) {
	testlib.Mktmp(t)
	testlib.GitInit(t)
	testlib.GitCommit(t, "first")
	ctx := context.New(config.Project{})
	require.Error(t, Pipe{}.Run(ctx))
	require.Empty(t, ctx.ReleaseNotes)
}

func TestChangelogOnBranchWithSameNameAsTag(t *testing.T) {
	testlib.Mktmp(t)
	testlib.GitInit(t)
	msgs := []string{
		"initial commit",
		"another one",
		"one more",
		"and finally this one",
	}
	for _, msg := range msgs {
		testlib.GitCommit(t, msg)
	}
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCheckoutBranch(t, "v0.0.1")
	ctx := context.New(config.Project{})
	ctx.Git.CurrentTag = "v0.0.1"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	for _, msg := range msgs {
		require.Contains(t, ctx.ReleaseNotes, msg)
	}
}

func TestChangeLogWithReleaseHeader(t *testing.T) {
	current, err := os.Getwd()
	require.NoError(t, err)
	tmpdir := testlib.Mktmp(t)
	require.NoError(t, os.Symlink(current+"/testdata", tmpdir+"/testdata"))
	testlib.GitInit(t)
	msgs := []string{
		"initial commit",
		"another one",
		"one more",
		"and finally this one",
	}
	for _, msg := range msgs {
		testlib.GitCommit(t, msg)
	}
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCheckoutBranch(t, "v0.0.1")
	ctx := context.New(config.Project{})
	ctx.Git.CurrentTag = "v0.0.1"
	ctx.ReleaseHeaderFile = "testdata/release-header.md"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	require.Contains(t, ctx.ReleaseNotes, "test header")
}

func TestChangeLogWithTemplatedReleaseHeader(t *testing.T) {
	current, err := os.Getwd()
	require.NoError(t, err)
	tmpdir := testlib.Mktmp(t)
	require.NoError(t, os.Symlink(current+"/testdata", tmpdir+"/testdata"))
	testlib.GitInit(t)
	msgs := []string{
		"initial commit",
		"another one",
		"one more",
		"and finally this one",
	}
	for _, msg := range msgs {
		testlib.GitCommit(t, msg)
	}
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCheckoutBranch(t, "v0.0.1")
	ctx := context.New(config.Project{})
	ctx.Git.CurrentTag = "v0.0.1"
	ctx.ReleaseHeaderTmpl = "testdata/release-header-templated.md"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	require.Contains(t, ctx.ReleaseNotes, "test header with tag v0.0.1")
}

func TestChangeLogWithReleaseFooter(t *testing.T) {
	current, err := os.Getwd()
	require.NoError(t, err)
	tmpdir := testlib.Mktmp(t)
	require.NoError(t, os.Symlink(current+"/testdata", tmpdir+"/testdata"))
	testlib.GitInit(t)
	msgs := []string{
		"initial commit",
		"another one",
		"one more",
		"and finally this one",
	}
	for _, msg := range msgs {
		testlib.GitCommit(t, msg)
	}
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCheckoutBranch(t, "v0.0.1")
	ctx := context.New(config.Project{})
	ctx.Git.CurrentTag = "v0.0.1"
	ctx.ReleaseFooterFile = "testdata/release-footer.md"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	require.Contains(t, ctx.ReleaseNotes, "test footer")
	require.Equal(t, rune(ctx.ReleaseNotes[len(ctx.ReleaseNotes)-1]), '\n')
}

func TestChangeLogWithTemplatedReleaseFooter(t *testing.T) {
	current, err := os.Getwd()
	require.NoError(t, err)
	tmpdir := testlib.Mktmp(t)
	require.NoError(t, os.Symlink(current+"/testdata", tmpdir+"/testdata"))
	testlib.GitInit(t)
	msgs := []string{
		"initial commit",
		"another one",
		"one more",
		"and finally this one",
	}
	for _, msg := range msgs {
		testlib.GitCommit(t, msg)
	}
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCheckoutBranch(t, "v0.0.1")
	ctx := context.New(config.Project{})
	ctx.Git.CurrentTag = "v0.0.1"
	ctx.ReleaseFooterTmpl = "testdata/release-footer-templated.md"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	require.Contains(t, ctx.ReleaseNotes, "test footer with tag v0.0.1")
	require.Equal(t, rune(ctx.ReleaseNotes[len(ctx.ReleaseNotes)-1]), '\n')
}

func TestChangeLogWithoutReleaseFooter(t *testing.T) {
	current, err := os.Getwd()
	require.NoError(t, err)
	tmpdir := testlib.Mktmp(t)
	require.NoError(t, os.Symlink(current+"/testdata", tmpdir+"/testdata"))
	testlib.GitInit(t)
	msgs := []string{
		"initial commit",
		"another one",
		"one more",
		"and finally this one",
	}
	for _, msg := range msgs {
		testlib.GitCommit(t, msg)
	}
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCheckoutBranch(t, "v0.0.1")
	ctx := context.New(config.Project{})
	ctx.Git.CurrentTag = "v0.0.1"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	require.Equal(t, rune(ctx.ReleaseNotes[len(ctx.ReleaseNotes)-1]), '\n')
}

func TestGetChangelogGitHub(t *testing.T) {
	ctx := context.New(config.Project{
		Changelog: config.Changelog{
			Use: "github",
		},
	})

	expected := "c90f1085f255d0af0b055160bfff5ee40f47af79: fix: do not skip any defaults (#2521) (@caarlos0)"
	mock := client.NewMock()
	mock.Changes = expected
	l := scmChangeloger{
		client: mock,
		repo: client.Repo{
			Owner: "goreleaser",
			Name:  "goreleaser",
		},
	}
	log, err := l.Log(ctx, "v0.180.1", "v0.180.2")
	require.NoError(t, err)
	require.Equal(t, expected, log)
}

func TestGetChangelogGitHubNative(t *testing.T) {
	ctx := context.New(config.Project{
		Changelog: config.Changelog{
			Use: "github-native",
		},
	})

	expected := "**Full Changelog**: https://github.com/gorelease/goreleaser/compare/v0.180.1...v0.180.2"
	mock := client.NewMock()
	mock.ReleaseNotes = expected
	l := githubNativeChangeloger{
		client: mock,
		repo: client.Repo{
			Owner: "goreleaser",
			Name:  "goreleaser",
		},
	}
	log, err := l.Log(ctx, "v0.180.1", "v0.180.2")
	require.NoError(t, err)
	require.Equal(t, expected, log)
}

func TestGetChangeloger(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		c, err := getChangeloger(context.New(config.Project{}))
		require.NoError(t, err)
		require.IsType(t, c, gitChangeloger{})
	})

	t.Run("git", func(t *testing.T) {
		c, err := getChangeloger(context.New(config.Project{
			Changelog: config.Changelog{
				Use: "git",
			},
		}))
		require.NoError(t, err)
		require.IsType(t, c, gitChangeloger{})
	})

	t.Run("github", func(t *testing.T) {
		ctx := context.New(config.Project{
			Changelog: config.Changelog{
				Use: "github",
			},
		})
		ctx.TokenType = context.TokenTypeGitHub
		c, err := getChangeloger(ctx)
		require.NoError(t, err)
		require.IsType(t, c, &scmChangeloger{})
	})

	t.Run("github-native", func(t *testing.T) {
		ctx := context.New(config.Project{
			Changelog: config.Changelog{
				Use: "github-native",
			},
		})
		ctx.TokenType = context.TokenTypeGitHub
		c, err := getChangeloger(ctx)
		require.NoError(t, err)
		require.IsType(t, c, &githubNativeChangeloger{})
	})

	t.Run("gitlab", func(t *testing.T) {
		ctx := context.New(config.Project{
			Changelog: config.Changelog{
				Use: "gitlab",
			},
		})
		ctx.TokenType = context.TokenTypeGitLab
		c, err := getChangeloger(ctx)
		require.NoError(t, err)
		require.IsType(t, c, &scmChangeloger{})
	})

	t.Run("invalid", func(t *testing.T) {
		c, err := getChangeloger(context.New(config.Project{
			Changelog: config.Changelog{
				Use: "nope",
			},
		}))
		require.EqualError(t, err, `invalid changelog.use: "nope"`)
		require.Nil(t, c)
	})
}

func TestSkip(t *testing.T) {
	t.Run("skip on snapshot", func(t *testing.T) {
		ctx := context.New(config.Project{})
		ctx.Snapshot = true
		require.True(t, Pipe{}.Skip(ctx))
	})

	t.Run("skip", func(t *testing.T) {
		ctx := context.New(config.Project{
			Changelog: config.Changelog{
				Skip: true,
			},
		})
		require.True(t, Pipe{}.Skip(ctx))
	})

	t.Run("dont skip", func(t *testing.T) {
		ctx := context.New(config.Project{})
		require.False(t, Pipe{}.Skip(ctx))
	})
}

func TestGroup(t *testing.T) {
	folder := testlib.Mktmp(t)
	testlib.GitInit(t)
	testlib.GitCommit(t, "first")
	testlib.GitTag(t, "v0.0.1")
	testlib.GitCommit(t, "added feature 1")
	testlib.GitCommit(t, "fixed bug 2")
	testlib.GitCommit(t, "ignored: whatever")
	testlib.GitCommit(t, "fix: whatever")
	testlib.GitCommit(t, "docs: whatever")
	testlib.GitCommit(t, "chore: something about cArs we dont need")
	testlib.GitCommit(t, "feat: added that thing")
	testlib.GitCommit(t, "bug: Merge pull request #999 from goreleaser/some-branch")
	testlib.GitCommit(t, "this is not a Merge pull request")
	testlib.GitTag(t, "v0.0.2")
	ctx := context.New(config.Project{
		Dist: folder,
		Changelog: config.Changelog{
			Groups: []config.ChangeLogGroup{
				{
					Title:  "Features",
					Regexp: "^.*feat[(\\w)]*:+.*$",
					Order:  0,
				},
				{
					Title:  "Bug Fixes",
					Regexp: "^.*bug[(\\w)]*:+.*$",
					Order:  1,
				},
				{
					Title:  "Catch nothing",
					Regexp: "yada yada yada honk the planet",
					Order:  10,
				},
				{
					Title: "Others",
					Order: 999,
				},
			},
		},
	})
	ctx.Git.CurrentTag = "v0.0.2"
	require.NoError(t, Pipe{}.Run(ctx))
	require.Contains(t, ctx.ReleaseNotes, "## Changelog")
	require.Contains(t, ctx.ReleaseNotes, "### Features")
	require.Contains(t, ctx.ReleaseNotes, "### Bug Fixes")
	require.NotContains(t, ctx.ReleaseNotes, "### Catch nothing")
	require.Contains(t, ctx.ReleaseNotes, "### Others")
}

func TestGroupBadRegex(t *testing.T) {
	folder := testlib.Mktmp(t)
	testlib.GitInit(t)
	testlib.GitCommit(t, "first")
	testlib.GitTag(t, "v0.0.1")
	testlib.GitTag(t, "v0.0.2")
	ctx := context.New(config.Project{
		Dist: folder,
		Changelog: config.Changelog{
			Groups: []config.ChangeLogGroup{
				{
					Title:  "Something",
					Regexp: "^.*feat[(\\w", // unterminated regex
				},
			},
		},
	})
	ctx.Git.CurrentTag = "v0.0.2"
	require.EqualError(t, Pipe{}.Run(ctx), `failed to group into "Something": error parsing regexp: missing closing ]: `+"`"+`[(\w`+"`")
}
