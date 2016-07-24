package cmd

import (
	"fmt"
	"strings"

	"github.com/smira/aptly/query"
	"github.com/smira/commander"
	"github.com/smira/flag"
)

func aptlyMirrorEdit(cmd *commander.Command, args []string) error {
	var err error
	if len(args) != 1 {
		cmd.Usage()
		return commander.ErrCommandError
	}

	repo, err := context.CollectionFactory().RemoteRepoCollection().ByName(args[0])
	if err != nil {
		return fmt.Errorf("unable to edit: %s", err)
	}

	err = repo.CheckLock()
	if err != nil {
		return fmt.Errorf("unable to edit: %s", err)
	}

	context.Flags().Visit(func(flag *flag.Flag) {
		if err != nil {
			return
		}

		switch flag.Name {
		case "filter":
			repo.Filter = flag.Value.String()
		case "filter-with-deps":
			repo.FilterWithDeps = flag.Value.Get().(bool)
		case "with-sources":
			repo.DownloadSources = flag.Value.Get().(bool)
		case "with-udebs":
			repo.DownloadUdebs = flag.Value.Get().(bool)
		case "with-repo-packages":
			repo.PackagesFromRepos = nil
			if queryRepos := flag.Value.String(); queryRepos != "" {
				queryRepoNames := strings.Split(queryRepos, ",")
				for _, queryRepoName := range queryRepoNames {
					queryRepo, err := context.CollectionFactory().LocalRepoCollection().ByName(queryRepoName)
					if err != nil {
						return
					}
					repo.PackagesFromRepos = append(repo.PackagesFromRepos, queryRepo.UUID)
				}
			}
		case "with-mirror-packages":
			repo.PackagesFromMirrors = nil
			if queryRepos := flag.Value.String(); queryRepos != "" {
				queryRepoNames := strings.Split(queryRepos, ",")
				for _, queryRepoName := range queryRepoNames {
					queryRepo, err := context.CollectionFactory().RemoteRepoCollection().ByName(queryRepoName)
					if err != nil {
						return
					}
					repo.PackagesFromMirrors = append(repo.PackagesFromMirrors, queryRepo.UUID)
				}
			}
		}
	})

	if err != nil {
		return fmt.Errorf("unable to edit: %s", err)
	}

	if repo.IsFlat() && repo.DownloadUdebs {
		return fmt.Errorf("unable to edit: flat mirrors don't support udebs")
	}

	if repo.Filter != "" {
		_, err = query.Parse(repo.Filter)
		if err != nil {
			return fmt.Errorf("unable to edit: %s", err)
		}
	}

	if context.GlobalFlags().Lookup("architectures").Value.String() != "" {
		repo.Architectures = context.ArchitecturesList()

		err = repo.Fetch(context.Downloader(), nil)
		if err != nil {
			return fmt.Errorf("unable to edit: %s", err)
		}
	}

	err = context.CollectionFactory().RemoteRepoCollection().Update(repo)
	if err != nil {
		return fmt.Errorf("unable to edit: %s", err)
	}

	fmt.Printf("Mirror %s successfully updated.\n", repo)
	return err
}

func makeCmdMirrorEdit() *commander.Command {
	cmd := &commander.Command{
		Run:       aptlyMirrorEdit,
		UsageLine: "edit <name>",
		Short:     "edit mirror settings",
		Long: `
Command edit allows one to change settings of mirror:
filters, list of architectures.

Example:

  $ aptly mirror edit -filter=nginx -filter-with-deps some-mirror
`,
		Flag: *flag.NewFlagSet("aptly-mirror-edit", flag.ExitOnError),
	}

	cmd.Flag.String("filter", "", "filter packages in mirror")
	cmd.Flag.Bool("filter-with-deps", false, "when filtering, include dependencies of matching packages as well")
	cmd.Flag.Bool("with-sources", false, "download source packages in addition to binary packages")
	cmd.Flag.Bool("with-udebs", false, "download .udeb packages (Debian installer support)")
	cmd.Flag.String("with-repo-packages", "", "when filtering, include packages contained in repos (comma separated)")
	cmd.Flag.String("with-mirror-packages", "", "when filtering, include packages contained in mirrors (comma separated)")
	return cmd
}
