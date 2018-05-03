#include "aes_seq.h"
#include <time.h>


long nonce = 0xaaaaaaaaaaaaaaaa;
clock_t start, end;
double cpu_time_used;
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
  if (b == 1) return a;
  if (b == 2) return gal2[a];
  if (b == 3) return gal3[a];
  
  return 0;
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
		unsigned char* temp = (unsigned char*) calloc(1, 4);
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

void xor (unsigned char* a, unsigned char* b) {
	addRoundKey(a, b);
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
		unsigned char* col = (unsigned char*) calloc(1, 4);
		for (int y = x * 4, z = 0; z < 4; y ++, z++) {
			col[z] = state[y];
		}

		mixSingleColumn(col);

		for (int y = x * 4, z = 0; z < 4; y ++, z++)
			state[y] = col[z];

    	free(col);
	}
}

void encrypt(unsigned char* state, unsigned char* expandedKeys, long cVal) {

	long* counter = (long*) calloc(sizeof(long), 2);
	counter[1] = nonce;
	counter[0] = cVal;

	unsigned char* counterState = (unsigned char*) counter;

	addRoundKey(counterState, expandedKeys);

  	for (int x = 1; x < 11; x++) {
  		subBytes(counterState);
  		shiftRows(counterState);

  		if (x != 10) /* no mixCols on last step */
  			mixColumns(counterState, 4);

  		addRoundKey(counterState, expandedKeys + 16 * x);
  	}

  	xor(state, counterState);
  	free(counterState);
}

void runAES(unsigned char* state, int stateLength, int blockSize, unsigned char* key) {
	/* to get correct size for padded buffer */
	int numBlocks = (stateLength / blockSize) + (stateLength % blockSize != 0);
	int bufferSize = numBlocks * blockSize;

	unsigned char* result = (unsigned char*) calloc(1,  bufferSize);

	int x;
	for (x = 0; x < stateLength; x++)
		result[x] = state[x];

	/* we need even block sizes, so we will pad any uneven block */
	char diff = blockSize - (stateLength % blockSize);
	while (x < bufferSize) {
		result[x] = diff;
		x++;
	}

	unsigned char* expandedKeys = (unsigned char*) calloc(1, 176);

	/* Key expansion - once */
  	keyExpansion(key, expandedKeys, 16, 176);

  	start = clock();
  	for (int x = 0; x < bufferSize; x += blockSize) {
  		encrypt(result + x, expandedKeys, (x / blockSize));
  	}
  	end = clock();
	cpu_time_used = ((double) (end - start)) / CLOCKS_PER_SEC;

	printf("%lf\n", 1000.0 * cpu_time_used);

  	// for (int x = 0; x < bufferSize; x++)
  	// 	printf("%x\n", result[x]);
}


char* read_input(char* input) {
	/* Open file and read in first 2 things*/
	FILE* input_file = fopen(input, "r");

	char* result;
	if (input_file != NULL) {
		fseek(input_file, 0L, SEEK_END);
		long s = ftell(input_file);
		rewind(input_file);

		result = malloc(s);
		if (result != NULL) {
			fread(result, s, 1, input_file);
			fclose(input_file);
			input_file = NULL;
		}
	}


	if (input_file != NULL) fclose(input_file);

	int len = strlen(result);
	if( result[len-1] == '\n' )
	    result[len-1] = 0;

	return result;
}

int main (int argc, char** argv) {
	char* keyFile = argv[1];
	char* stateFile = argv[2];

	char* key = read_input(keyFile);
	char* state = read_input(stateFile);

  	runAES((unsigned char*) state, strlen(state), 16, (unsigned char*) key);
}
