package main

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"
)

var blockSize = 16

var mixColumnChan = make(chan *Params)

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

	output[0] = output[0] ^ rcon[iter]

	return output
}

func expandKey(key []byte, numExpandedBytes int) []byte {

	keyLen := len(key)

	var ret = make([]byte, numExpandedBytes)
	copy(ret[:keyLen], key[:keyLen])
	i := 1

	for currInd := keyLen; currInd < numExpandedBytes; {
		temp := make([]byte, 4)
		copy(temp[0:4], ret[currInd-4:currInd])
		temp = keySchedCore(temp, i)

		for k := 0; k < 4; k++ {
			temp[k] = temp[k] ^ ret[currInd-keyLen+k]
		}

		copy(ret[currInd:currInd+4], temp[0:4])
		currInd += 4

		for k := 0; k < 3; k++ {
			copy(temp[0:4], ret[currInd-4:currInd])
			for j := 0; j < 4; j++ {
				temp[j] = temp[j] ^ ret[currInd-keyLen+j]
			}
			copy(ret[currInd:currInd+4], temp[0:4])
			currInd += 4
		}
		i++
	}
	return ret
}

func subBytesGoRout(inputPtr *Params) {
	input := *inputPtr
	state := *(input.statePtr)

	for i := 0; i < len(state); i++ {
		state[i] = s_box[state[i]]
	}
}

func addRoundKey(inputPtr *Params) {
	input := *inputPtr
	state := *(input.statePtr)
	key := *(input.expandedKeyPtr)

	for i := 0; i < len(state); i++ {
		state[i] = state[i] ^ key[i]
	}
}

func leftRotateByOne(state []byte, row int, size int) {
	temp := state[row]
	for i := 0; i < size-1; i++ {
		cur := row
		next := row + 4
		state[cur] = state[next]
		row += 4
	}
	state[row] = temp
}

func shiftRows(inputPtr *Params) {
	input := *inputPtr
	state := *(input.statePtr)

	for i := 1; i <= 3; i++ {
		for k := 0; k < i; k++ {
			leftRotateByOne(state, i, 4)
		}
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

func mixColumns(inputPtr *Params) {
	input := *inputPtr
	state := *(input.statePtr)
	numCols := input.numCols

	for i := 0; i < numCols; i++ {
		col := make([]byte, 4)
		copy(col[0:4], state[(i*4):((i+1)*4)])
		mixSingleColumn(col)
		copy(state[(i*4):((i+1)*4)], col[0:4])
	}
}

func encrypt(nonce uint64, counter uint64, expandedKeyPtr *[]byte, plaintext []byte, wg *sync.WaitGroup) {

	state := make([]byte, blockSize)

	//https://stackoverflow.com/questions/35371385/how-can-i-convert-an-int64-into-a-byte-array-in-go

	binary.LittleEndian.PutUint64(state[:8], counter)
	binary.LittleEndian.PutUint64(state[8:], nonce)

	//channel to allow workers to notify corresponding encrypt routine when done
	c := make(chan bool)
	input := Params{statePtr: &state, expandedKeyPtr: expandedKeyPtr, numCols: 4, c: c}

	addRoundKey(&input)

	for i := 1; i < 11; i++ {

		subBytesGoRout(&input)
		shiftRows(&input)

		if i != 10 {
			mixColumns(&input)
		}
		temp := (*expandedKeyPtr)[blockSize*i : blockSize*(i+1)]
		input.expandedKeyPtr = &temp
		addRoundKey(&input)
	}

	for i := 0; i < len(plaintext); i++ {
		plaintext[i] = plaintext[i] ^ state[i]
	}
	(*wg).Done()
	close(c)

}

func main() {

	//cut off anything thats longer than the key
	keyBytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	keyBytes = keyBytes[0:16]

	state, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}

	//padding to the end by the number of padded bytes
	if len(state)%blockSize != 0 {
		diff := blockSize - (len(state) % blockSize)
		toAppend := make([]byte, diff)
		for i := 0; i < diff; i++ {
			toAppend[i] = byte(diff)
		}
		state = append(state, toAppend...)
	}

	expandedKey := expandKey(keyBytes, 176)
	expandedKeyPtr := &expandedKey

	var nonce uint64 = 0xAAAAAAAAAAAAAAAA
	counter := 0

	var wg sync.WaitGroup

	start := time.Now()

	for i := 0; i < len(state); i += blockSize {
		go encrypt(nonce, uint64(counter), expandedKeyPtr, state[i:i+blockSize], &wg)
		counter++
		wg.Add(1)
	}
	wg.Wait()

	end := time.Now()

	fmt.Printf("%d\n", end.Sub(start).Nanoseconds())

	// for i := 0; i < len(state); i++ {
	//   fmt.Printf("%x ", state[i])
	// }
}
