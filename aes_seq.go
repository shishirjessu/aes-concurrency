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

func keySchedCore(word []byte, iter int) []byte{
  output := make([]byte, 4)
  copy(output[:4], word[:4])


  temp := output[0]
  output[0] = output[1]
  output[1] = output[2]
  output[2] = output[3]
  output[3] = temp

  for i := 0; i < 4; i++ {
    output[i] = s_box[output[i]]
  }

  output[0] = output[0]^rcon[iter]

  return output
}

func expandKey(key []byte, numExpandedBytes int) []byte {

  keyLen := len(key)


  var ret = make([]byte, numExpandedBytes)
  copy(ret[:keyLen], key[:keyLen])
  currInd := keyLen

  for i := 1; i < 11; i++ {
    temp := make([]byte, 4)
    copy(temp[0:4], ret[currInd-4: currInd])
    temp = keySchedCore(temp, i)

    for k := 0; k < 4; k++ {
      temp[k] = temp[k] ^ ret[currInd-keyLen+k]
    }

    copy(ret[currInd : currInd+4], temp[0:4])
    currInd += 4

    for k := 0; k < 3; k++ {
      copy(temp[0:4], ret[currInd-4: currInd])
      for j := 0; j < 4; j++ {
        temp[j] = temp[j] ^ ret[currInd-keyLen+j]
      }
      copy(ret[currInd: currInd+4], temp[0:4])
      currInd += 4
    }
  }
  print(ret)
  return ret

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
  // print(str)
  letters := make([]byte, len(str))
  for i := 0; i < len(str); i++ {
    letters[i] = str[i]
  }
  // printBlock(letters)
  // shiftRows(letters)
  // printBlock(letters)

  key := "1234567890123456"
  keyBytes := make([]byte, len(key))
  for i := 0; i < len(key); i++ {
    keyBytes[i] = key[i]
  }
  // print()
  // print (keyBytes)
  expandKey(keyBytes, 176)

  // print(s_box)


}
