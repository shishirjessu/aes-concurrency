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

func leftRotateByOne(state []byte, row int, size int)  {
  temp := state[row]
  for i := 0; i < size-1; i++ {
    cur := row
    next := row + 4
    state[cur] = state[next]
    row += 4
  }
  state[row] = temp
}

func shiftRows(state []byte) {
  for i := 1; i <= 3; i++ {
    for k := 0; k < i; k++ {
      leftRotateByOne(state, i, 4)
    }
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
  shiftRows(letters)
  printBlock(letters)


}
