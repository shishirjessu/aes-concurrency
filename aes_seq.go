package main

import (

  "fmt"

)

var print = fmt.Println

func printBlock (letters []byte) {
  for i := 0; i < 4; i++ {
    temp := ""
    for k := 0; k < 4; k++ {
      temp += string(letters[k*4 + i])
      temp += ""
    }
    print(temp)
  }
}

func main() {
  str := "How are u world?"
  print(str)
  letters := make([]byte, len(str))
  for i := 0; i < len(str); i++ {
    letters[i] = str[i]
  }
  printBlock(letters)


}
