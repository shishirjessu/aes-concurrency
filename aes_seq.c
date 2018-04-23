#include "aes_seq.h"


/* print out 16-byte block as grid */
void printBlock(unsigned char* state) {
	for (int x = 0; x < 4; x++) {
		for (int y = 0; y < 4; y++)
			printf("%x ", state[y * 4 + x]);

		printf("\n");
	}
}

/* left-shifts a row, data is is column order */
void leftRotateByOne(unsigned char* state, int row, int size) {
	char temp = state[row];
	int x;
	for (x = 0; x < size - 1; x++) {
		int cur = row;
		int next = row + 4;

		state[cur] = state[next];
		row += 4;
	}
	state[row] = temp;
}

unsigned char gmul (unsigned char a, unsigned char b) {
	unsigned char product = 0;

	for (int x = 0; x < 8; x++) {
		if (b & 0x1)
			product ^= a;

		int highBitSet = a & 0x80;
		a <<= 1;

		if (highBitSet == 0x80)
			a ^= 0x1b;

		b >>= 1;
	}

	return product;
}


void mixSingleColumn(unsigned char* col) {
	unsigned char temp[4];

	for (int x = 0; x < 4; x++)
		temp[x] = col[x];

	col[0] = gmul(temp[0], 2) ^ gmul(temp[3], 1) ^ gmul(temp[2], 1) ^ gmul(temp[1], 3);
	col[1] = gmul(temp[1], 2) ^ gmul(temp[0], 1) ^ gmul(temp[3], 1) ^ gmul(temp[2], 3);
	col[2] = gmul(temp[2], 2) ^ gmul(temp[1], 1) ^ gmul(temp[0], 1) ^ gmul(temp[3], 3);
	col[3] = gmul(temp[3], 2) ^ gmul(temp[2], 1) ^ gmul(temp[1], 1) ^ gmul(temp[0], 3);
}

/* the main steps of AES */

/* prepares one 4-byte word */
void keyScheduleCore (unsigned char* word, int round) {
	char temp = word[0];
	word[0] = word[1];
	word[1] = word[2];
	word[2] = word[3];
	word[3] = temp;

	for (int x = 0; x < 4; x++)
		word[x] = sub_bytes_lookup[word[x]];

	word[0] ^= rcon[round];
}


/* expand them keys brah
   n: number of bytes in the original key
   b: number of total bytes we want
*/
void keyExpansion (unsigned char* key, unsigned char* expandedKeys, int n, int b) {
	int numExp;
	for (numExp = 0; numExp < n; numExp++)
		expandedKeys[numExp] = key[numExp];

	int round = 1;
	while (numExp < b) {
		/* copy over the last 4 bytes of the expanded key to temp */
		unsigned char* temp = calloc(1, 4);
		for (int x = numExp - 4, y = 0; y < 4; x++, y++)
			temp[y] = expandedKeys[x];

		/* perform the core on temp, incrementing round when done */
		keyScheduleCore(temp, round++);

		/* x-or temp with whatever's n bytes before it, then drop it in expandedKeys*/
		for (int x = numExp - n, y = 0; y < 4; y++, x++)
			temp[y] ^= expandedKeys[x];

		for (int x = 0; x < 4; x++)
			expandedKeys[numExp + x] = temp[x];

		/* now we've expanded 4 more bytes */
		numExp += 4;

		/* need to produce the next 12 bytes of expanded key */
		for (int a = 0; a < 3; a++) {
			/* grab the previous 4 bytes and put it in temp */
			for (int x = numExp - 4, y = 0; y < 4; x++, y++)
				temp[y] = expandedKeys[x];

			/* x-or temp with whatever's n bytes before it, then drop it in expandedKeys*/
			for (int x = numExp - n, y = 0; y < 4; y++, x++)
				temp[y] ^= expandedKeys[x];

			for (int x = 0; x < 4; x++)
				expandedKeys[numExp + x] = temp[x];

			numExp += 4;
		}
    	free(temp);
	}
}

/* switch data with corresponding data in Rijndael s-box */
void subBytes (unsigned char* state) {
	for (int x = 0; x < 16; x++)
		state[x] = sub_bytes_lookup[state[x]];
}

void addRoundKey (unsigned char* state, unsigned char* key) {
	for (int x = 0; x < 16; x++)
		state[x] ^= key[x];
}

void shiftRows (unsigned char* state) {
	for (int x = 1; x <= 3; x++) {
		for (int y = 0; y < x; y++)
			leftRotateByOne(state, x, 4);
	}
}

/* perform the mix columns operation
   n: the number of columns */
void mixColumns (unsigned char* state, int n) {
	for (int x = 0; x < n; x++) {
		unsigned char* col = calloc(1, 4);
		for (int y = x * 4, z = 0; z < 4; y ++, z++) {
			col[z] = state[y];
		}

		mixSingleColumn(col);

		for (int y = x * 4, z = 0; z < 4; y ++, z++)
			state[y] = col[z];

    	free(col);
	}
}

int main (int argc, char** argv) {
  	unsigned char* str = "Two One Nine Two";
  	int len = 16;
	unsigned char* start = calloc(1, sizeof(char) * len);
  	memcpy(start, str, len * sizeof(char));

	unsigned char* key = "Thats my Kung Fu";
	unsigned char* keyBytes = calloc(1, sizeof(char) * len);
  	memcpy(keyBytes, key, len * sizeof(char));

  	unsigned char* expandedKeys = calloc(1, 176);

  	/* Key expansion - once */
  	keyExpansion(keyBytes, expandedKeys, 16, 176);

  	for (int x = 0; x < 11; x++) {
  		for (int y = 0; y < 16; y++)
  			printf("%x ", expandedKeys[x * 16 + y]);
  		printf("\n");
  	}
  	printf("\n");

  	addRoundKey(start, expandedKeys);

  	for (int x = 1; x < 11; x++) {
  		subBytes(start);
  		shiftRows(start);

  		if (x != 10) /* no mixCols on last step */
  			mixColumns(start, 4);

  		addRoundKey(start, expandedKeys + 16 * x);
  	} 

  	for (int x = 0; x < 16; x++) 
  		printf("%x ", start[x]);

	free(start);
  	free(keyBytes);
}
