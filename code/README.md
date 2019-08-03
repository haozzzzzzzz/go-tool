#The `code` is a CLI library for Go source code developing


```sh
code is a CLI library for Go source code developing.

Usage:
  code [command]

Available Commands:
  help        Help about any command
  precompile  precompile --path filepath [ --params_file params_filepath ] [ key1=val1 key2=val2 ]

Flags:
      --config string   config file (default is $HOME/.code.yaml)
  -h, --help            help for code
  -t, --toggle          Help message for toggle

Use "code [command] --help" for more information about a command.
```



## precompile

```sh
precompile --path filepath [ --params_file params_filepath ] [ key1=val1 key2=val2 ]

Usage:
  code precompile [flags]

Flags:
  -h, --help            help for precompile
      --params string   params file path
  -p, --path string     precompile path (default "./")

Global Flags:
      --config string   config file (default is $HOME/.code.yaml)
```



**How to use precompile**

- Naming your file that needs precompilation with suffix ".pre.go", like `xray.pre.go`.

- Add an `+pre` directive in comment line to declare an precompile statement.

- Add Lua syntax rules script to make precompile statement.

  ```go
  import (
  	"github.com/sirupsen/logrus"
  	"gopkg.in/go-playground/validator.v9"
  
  	"github.com/aws/aws-xray-sdk-go/xray"
  
  	// Importing the plugins enables collection of AWS resource information at runtime.
  	// Every plugin should be imported after "github.com/aws/aws-xray-sdk-go/xray" library.
  	// Precompiled by different host platform: ec2、ecs、beanstalk
  
  	// +pre if ( platform == "ec2" ) then
  	_ "github.com/aws/aws-xray-sdk-go/plugins/ec2"
  
  	// +pre elseif ( platform == "ecs" ) then
  	_ "github.com/aws/aws-xray-sdk-go/plugins/ecs"
  
  	// +pre elseif ( platform == "beanstalk" ) then
  	//_ "github.com/aws/aws-xray-sdk-go/plugins/beanstalk"
  
  	// +pre end
  
  	"github.com/aws/aws-xray-sdk-go/strategy/sampling"
  )
  ```

- Use cli cmd to precompile

  ```sh
  code precompile -p path_to_dir_or_file platform=ec2
  ```

- Or you can specify an params file path

  ```yaml
  // params.yaml
  platform=ecs
  ```

  ```sh
  code precompile -p path_to_dir_or_file --params path_to_params_file
  ```

- If you specify both params file and cmd args `[ --params_file params_filepath ] [ key1=val1 key2=val2 ]`,  they will by merged, and cmd args has a higher  priority then params file.

  ```sh
  code precompile -p path_to_dir_or_file --params path_to_params_file platform=ec2
  ```

  finally platform will be `ec2`

- Precompiled file will create an new file without ".pre.go" suffix.

  ```go
  //xray.go
  
  import (
  	"github.com/sirupsen/logrus"
  	"gopkg.in/go-playground/validator.v9"
  
  	"github.com/aws/aws-xray-sdk-go/xray"
  
  	// Importing the plugins enables collection of AWS resource information at runtime.
  	// Every plugin should be imported after "github.com/aws/aws-xray-sdk-go/xray" library.
  	// Precompiled by different host platform: ec2、ecs、beanstalk
  
  	_ "github.com/aws/aws-xray-sdk-go/plugins/ec2"
  
  	"github.com/aws/aws-xray-sdk-go/strategy/sampling"
  )
  ```

  