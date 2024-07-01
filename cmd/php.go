package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// phpCmd represents the php command
var phpCmd = &cobra.Command{
	Use:   "php",
	Short: "Switch to a specific version of php-cli",

	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			// `cw php` invoked (no args) so autodetect bin path
			return
		}

		newVersion := args[0]

		currentVersionCmd := exec.Command("php", "-v")
		versionOutput, err := currentVersionCmd.Output()

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		currentVersionOutput := string(versionOutput)
		currentVersion := currentVersionOutput[4:7]

		if newVersion == currentVersion {
			fmt.Println("You're already using PHP", newVersion)
			os.Exit(0)
		}

		brewPrefixCmd := exec.Command("brew", "--prefix")
		prefixOutput, err := brewPrefixCmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		brewPrefixPath := strings.TrimSpace(string(prefixOutput))

		phpVersions, err := os.ReadDir(brewPrefixPath + "/etc/php")
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		validVersion := false
		for _, dir := range phpVersions {
			if newVersion == dir.Name() {
				validVersion = true
			}
		}

		if validVersion {
			fmt.Printf("Switching PHP version %s -> %s\n", currentVersion, newVersion)

			unlinkCmd := exec.Command("brew", "unlink", "php@"+currentVersion)
			fmt.Printf("Command: [%s]\n", unlinkCmd.String())

			stdout, _ := unlinkCmd.StdoutPipe()
			stderr, _ := unlinkCmd.StderrPipe()

			_ = unlinkCmd.Start()

			scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				m := scanner.Text()
				fmt.Println(m)
			}

			_ = unlinkCmd.Wait()

			linkCmd := exec.Command("brew", "link", "php@"+newVersion, "--force", "--overwrite")
			if err := linkCmd.Run(); err != nil {
				log.Fatal(err)
			}

		} else {
			fmt.Println("PHP version", newVersion, "is not available.")
			os.Exit(1)
		}

	},
}

func init() {
	rootCmd.AddCommand(phpCmd)
}
