package source

import "os"

var ProjectDirMode = os.ModePerm ^ 0111 // 0666
