package compile

import (
	"github.com/haozzzzzzzz/go-tool/lib/lswagger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func CmdMergeSwagger() (cmd *cobra.Command) {
	var baseFile string
	var appendFile string
	var outputFile string
	var pretty bool
	cmd = &cobra.Command{
		Use: "swagger_merge",
		Run: func(cmd *cobra.Command, args []string) {
			if baseFile == "" || appendFile == "" || outputFile == "" {
				logrus.Errorf("invalid params")
				return
			}

			err := merge(baseFile, appendFile, outputFile, pretty)
			if err != nil {
				logrus.Errorf("merge failed. error: %s", err)
				return
			}
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&baseFile, "base", "b", "", "base file")
	flags.StringVarP(&appendFile, "append", "a", "", "append file")
	flags.StringVarP(&outputFile, "output", "o", "", "output file")
	flags.BoolVarP(&pretty, "pretty", "P", false, "pretty swagger.json output")

	return
}

func merge(baseFile string, appendFile string, outputFile string, pretty bool) (err error) {
	baseSwagger, err := lswagger.LoadSwagger(baseFile)
	if err != nil {
		logrus.Errorf("load swagger failed. error: %s", err)
		return
	}

	appendSwagger, err := lswagger.LoadSwagger(appendFile)
	if err != nil {
		logrus.Errorf("load swagger failed. error: %s", err)
		return
	}

	baseSwagger.Merge(appendSwagger)

	err = baseSwagger.SaveFile(outputFile, pretty)
	if err != nil {
		logrus.Errorf("save file failed. error: %s", err)
		return
	}

	return
}
