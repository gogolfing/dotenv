package dotenv

import (
	"fmt"
	"os"
	"strings"
)

func Example() {
	source := `
name1=value1
export name2=value2
name3="Hello\nWorld"

name4=value4 # this is a comment

# this is a comment as well
`

	sourcer := NewSourcer()

	err := sourcer.Source(strings.NewReader(source))

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(os.Getenv("name1"))
	fmt.Println(os.Getenv("name2"))
	fmt.Println(os.Getenv("name3"))
	fmt.Println(os.Getenv("name4"))
	//Output:
	//value1
	//value2
	//Hello
	//World
	//value4
}
