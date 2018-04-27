package main

import (
  "fmt"
)

var print = fmt.Println

func printBlock(letters []byte) {
  for i := 0; i < 4; i++ {
    temp := ""
    for k := 0; k < 4; k++ {
      temp += string(letters[k*4 + i])
      temp += ""
    }
    print(temp)
  }
}

func subBytes(input []byte) {
  for i := 0; i < len(input); i++ {
    input[i] = s_box[input[i]]
  }
}

func gmul(a byte, b byte) byte {
  if b == 1 {
    return a
  }
  if b == 2 {
    return gal2[a]
  }
  if b == 3 {
    return gal3[a]
  }
  return 0
}

func mixSingleColumn(column []byte) {

  temp := make([]byte, 4)
  copy(temp[0:4], column[0:4])

  column[0] = gmul(temp[0], 2) ^ gmul(temp[3], 1) ^ gmul(temp[2], 1) ^ gmul(temp[1], 3)
  column[1] = gmul(temp[1], 2) ^ gmul(temp[0], 1) ^ gmul(temp[3], 1) ^ gmul(temp[2], 3)
  column[2] = gmul(temp[2], 2) ^ gmul(temp[1], 1) ^ gmul(temp[0], 1) ^ gmul(temp[3], 3)
  column[3] = gmul(temp[3], 2) ^ gmul(temp[2], 1) ^ gmul(temp[1], 1) ^ gmul(temp[0], 3)

}

func mixColumns(state []byte, numCols int) {
  for i := 0; i < numCols; i++ {
    col := make([]byte, 4)
    copy(col[0:4], state[(i*4):((i+1)*4)])

    mixSingleColumn(col)

    copy(state[(i*4):((i+1)*4)], col[0:4])
  }
}

func keySchedCore(word []byte, iter int) []byte {
  output := make([]byte, 4)
  copy(output[:4], word[:4])

  temp := output[0]
  output[0] = output[1]
  output[1] = output[2]
  output[2] = output[3]
  output[3] = temp

  subBytes(output)

  output[0] = output[0]^rcon[iter]

  return output
}

func expandKey(key []byte, numExpandedBytes int) []byte {

  keyLen := len(key)

  var ret = make([]byte, numExpandedBytes)
  copy(ret[:keyLen], key[:keyLen])
  i := 1

  for currInd := keyLen; currInd < numExpandedBytes; {
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
    i++
  }
  // print(ret)
  return ret
}

func addRoundKey(state []byte, key []byte) {
  for i := 0; i < len(state); i++ {
    state[i] = state[i]^key[i]
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

func encrypt(state []byte, expandedKey []byte)  {
  addRoundKey(state, expandedKey)

  for i := 1; i < 11; i++ {
    subBytes(state)
    shiftRows(state)

    if i != 10 {
      mixColumns(state, 4)
    }
    addRoundKey(state, expandedKey[16*i:])

  }
}

func main() {
  str := "Two One Nine Two xyz123"
  state := make([]byte, len(str))
  for i := 0; i < len(str); i++ {
    state[i] = str[i]
  }

  if len(state) % 16 != 0 {
    diff := 16 - (len(state) % 16)
    toAppend := make([]byte, diff)
    for i := 0; i < diff; i++ {
      toAppend[i] = byte(diff)
    }
    state = append(state, toAppend...)
  }

  key := "Thats my Kung Fu"
  keyBytes := make([]byte, len(key))
  for i := 0; i < len(key); i++ {
    keyBytes[i] = key[i]
  }

  expandedKey := expandKey(keyBytes, 176)

  for i := 0; i < len(state); i += 16 {
    encrypt(state[i:i+16], expandedKey)
  }

  for i := 0; i < len(state); i++ {
    fmt.Printf("%x ", state[i])
  }


}
