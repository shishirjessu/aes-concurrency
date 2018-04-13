#include "aes_seq.h"
#include <stdio.h>
#include <string.h>


/* print out 16-byte block as grid */
void printBlock(unsigned char* state) {
	for (int x = 0; x < 4; x++) {
		for (int y = 0; y < 4; y++)
			printf("%c ", state[y * 4 + x]);

		printf("\n");
	}
}

/* the main steps of AES */
void keyExpansion () {

}

/* switch data with corresponding data in Rijndael s-box */
void subBytes (unsigned char* state) {
	for (int x = 0; x < 16; x++)
		state[x] = sub_bytes_lookup[state[x]];
}

void addRoundKey () {

}

/* left-shifts a row, data is is column order */
void leftRotateByOne(unsigned char* state, int row, int size) {
	char temp = state[row];
	int x;
	for (x = 0; x < size - 1; x++) {
		int cur = row;
		int next = row + 4;
		// printf("%d %d %d\n", cur, next, row);
		state[cur] = state[next];
		row += 4;
	}
  printf("\n");
	state[row] = temp;
}

void shiftRows (unsigned char* state) {
	for (int x = 1; x <= 3; x++) {
		for (int y = 0; y < x; y++)
			leftRotateByOne(state, x, 4);
	}
}

void mixColumns () {

}

int main (int argc, char** argv) {
  char *str = "How are u world?";
	unsigned char* start = calloc(1, sizeof(char) * (strlen(str)));
  memcpy(start, str, strlen(str) * sizeof(char));
	printBlock(start);
	shiftRows(start);
	printBlock(start);
  free(start);
}
