package main

import (
  "fmt"
  "encoding/binary"
  "sync"
)

var print = fmt.Println
var blockSize = 16

var mixColumnChan = make(chan *Params)
var subBytesChan = make(chan *Params)
var shiftRowsChan = make(chan *Params)
var addRoundKeyChan = make(chan *Params)

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

func subBytes(statePtr *[]byte) {
  state := *statePtr
  for i := 0; i < len(state); i++ {
    state[i] = s_box[state[i]]
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

  subBytes(&output)

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
  return ret
}

func subBytesGoRout() {
  for ;true; {
    input := <-subBytesChan
    state := *(input.statePtr)
    c := input.c

    for i := 0; i < len(state); i++ {
      state[i] = s_box[state[i]]
    }
    c <- true
  }
}

func addRoundKey() {
  for ;true; {
    input := <- addRoundKeyChan

    state := *(input.statePtr)
    key := *(input.expandedKeyPtr)
    c := input.c

    for i := 0; i < len(state); i++ {
      state[i] = state[i]^key[i]
    }
    c <- true
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

func shiftRows() {
  for ;true; {
    input := <-shiftRowsChan
    state := *(input.statePtr)
    c := input.c

    for i := 1; i <= 3; i++ {
      for k := 0; k < i; k++ {
        leftRotateByOne(state, i, 4)
      }
    }
    c <- true
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

func mixColumns() {
  for ;true; {
    input := <-mixColumnChan

    state := *(input.statePtr)
    numCols := input.numCols
    c := input.c

    for i := 0; i < numCols; i++ {
      col := make([]byte, 4)
      copy(col[0:4], state[(i*4):((i+1)*4)])
      mixSingleColumn(col)
      copy(state[(i*4):((i+1)*4)], col[0:4])
    }
    c <- true
  }
}


func encrypt(nonce uint64, counter uint64, expandedKey []byte, plaintext []byte, wg
  *sync.WaitGroup)  {

  state := make([]byte, blockSize)

  binary.LittleEndian.PutUint64(state[:8], counter)
  binary.LittleEndian.PutUint64(state[8:], nonce)

  c := make(chan bool)
  input := Params{statePtr:&state, expandedKeyPtr:&expandedKey, numCols:4, c:c}

  addRoundKeyChan <- &input
  _ = <-c

  for i := 1; i < 11; i++ {

    subBytesChan <- &input
    _ = <-c

    shiftRowsChan <- &input
    _ = <-c

    if i != 10 {
      mixColumnChan <- &input
      _ = <-c
    }
    temp := expandedKey[blockSize*i:blockSize*(i+1)]
    input.expandedKeyPtr = &temp
    addRoundKeyChan <- &input
    _ = <-c
  }

  for i := 0; i < len(plaintext); i++ {
    plaintext[i] = plaintext[i]^state[i]
  }
  (*wg).Done()

}

func main() {
  str := "Two One Nine Two xyz123"
  state := make([]byte, len(str))
  for i := 0; i < len(str); i++ {
    state[i] = str[i]
  }

  if len(state) % blockSize != 0 {
    diff := blockSize - (len(state) % blockSize)
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

  for i := 0; i < 2; i++ {
    go subBytesGoRout()
    go shiftRows()
    go mixColumns()
    go addRoundKey()
  }

  var nonce uint64 = 0xAAAAAAAAAAAAAAAA
  counter := 0

  var wg sync.WaitGroup

  for i := 0; i < len(state); i += blockSize {
    go encrypt(nonce, uint64(counter), expandedKey, state[i:i+blockSize], &wg)
    counter++
    wg.Add(1)
  }
  wg.Wait()

  for i := 0; i < len(state); i++ {
    fmt.Printf("%x ", state[i])
  }
}
