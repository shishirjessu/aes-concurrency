#include "aes-ctr.cuh"
#include "aes_seq.h"

#define MAX_THREADS 1024

void err_chk() {
	cudaDeviceSynchronize();
	cudaError_t err = cudaGetLastError();
	if (err != cudaSuccess)
		printf("Error: %s\n", cudaGetErrorString(err));
	// else
	// 	printf("%s\n", "cudaSuccess");
}

/* print out 16-byte block as grid */
__device__ void printBlock(unsigned char* state) {
	for (int x = 0; x < 4; x++) {
		for (int y = 0; y < 4; y++)
			printf("%x ", state[y * 4 + x]);

		printf("\n");
	}
}

/* left-shifts a row, data is is column order */
__device__ void leftRotateByOne(unsigned char* state, int row, int size) {
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

__device__ unsigned char gmul (unsigned char a, unsigned char b) {
  if (b == 1) return a;
  if (b == 2) return d_gal2[a];
  if (b == 3) return d_gal3[a];

  return 0;
}


__device__ void mixSingleColumn(unsigned char* col) {
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
		unsigned char temp[4] = {0, 0, 0, 0};
		//unsigned char* temp = (unsigned char*) calloc(1, 4);
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
	}
}

/* switch data with corresponding data in Rijndael s-box */
__device__ void subBytes (unsigned char* state) {
	for (int x = 0; x < 16; x++)
		state[x] = d_sub_bytes_lookup[state[x]];
}

__device__ void addRoundKey (unsigned char* state, unsigned char* key) {
	for (int x = 0; x < 16; x++)
		state[x] ^= key[x];
}

__device__ void xOr (unsigned char* a, unsigned char* b) {
	addRoundKey(a, b);
}

__device__ void shiftRows (unsigned char* state) {
	for (int x = 1; x <= 3; x++) {
		for (int y = 0; y < x; y++)
			leftRotateByOne(state, x, 4);
	}
}

/* perform the mix columns operation
   n: the number of columns */
__device__ void mixColumns (unsigned char* state, int n) {
	for (int x = 0; x < n; x++) {
		unsigned char col [4] = {0, 0, 0, 0};
		//unsigned char* col = (unsigned char*) calloc(1, 4);
		for (int y = x * 4, z = 0; z < 4; y ++, z++) {
			col[z] = state[y];
		}

		mixSingleColumn(col);

		for (int y = x * 4, z = 0; z < 4; y ++, z++)
			state[y] = col[z];

    	free(col);
	}
}

__global__ void encrypt(unsigned char* state, unsigned char* expandedKeys, int bufferSize) {
	long cVal = blockIdx.x*blockDim.x+threadIdx.x;
	__syncthreads();
	// if (cVal == 0) {
	// 	for (int x = 0; x < bufferSize; x++)
 //  			printf("%x\n", state[x]);
	// }
	__syncthreads();

	long nonce = 0xaaaaaaaaaaaaaaaa;

	long counter [2];

	counter[0] = cVal;
	counter[1] = nonce;

	unsigned char* counterState = (unsigned char*) counter;

	addRoundKey(counterState, expandedKeys);

	for (int x = 1; x < 11; x++) {
		subBytes(counterState);
		shiftRows(counterState);

		if (x != 10)
			mixColumns(counterState, 4);

		addRoundKey(counterState, expandedKeys + 16 * x);
	}

	int blockSize = 16;
	unsigned char* toXor = state + (blockSize * cVal);

	xOr(toXor, counterState);

	__syncthreads();
	// if (cVal == 0) {
	// 	for (int x = 0; x < bufferSize; x++)
 //  			printf("%d %x\n", x, state[x]);
	// }
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

	int numExpandedKeyBytes = 176;

	unsigned char* expandedKeys = (unsigned char*) calloc(1, numExpandedKeyBytes);

	/* Key expansion - once */
  	keyExpansion(key, expandedKeys, 16, numExpandedKeyBytes);

  	/* Throw stuff onto GPU */
  	unsigned char *dState, *dExpandedKeys; /* device ptrs */
  	cudaMalloc((void**)&dState, bufferSize);
  	err_chk();

  	cudaMemcpy(dState, result, bufferSize, cudaMemcpyHostToDevice);
  	err_chk();

  	cudaMalloc((void**)&dExpandedKeys, numExpandedKeyBytes);
  	err_chk();

  	cudaMemcpy(dExpandedKeys, expandedKeys, numExpandedKeyBytes, cudaMemcpyHostToDevice);
  	err_chk();

  	int numThreadBlocks = (numBlocks + MAX_THREADS - 1) / MAX_THREADS;
  	int threadsPerBlock = numBlocks / numThreadBlocks;

  	// printf("num threads: %d %d %d\n", numThreadBlocks, threadsPerBlock, numThreadBlocks * threadsPerBlock);

  	cudaEvent_t start, stop;
	cudaEventCreate(&start);
	cudaEventCreate(&stop);
	cudaEventRecord(start);

  	encrypt <<<numThreadBlocks, threadsPerBlock>>> (dState, dExpandedKeys, bufferSize);
  	err_chk();

  	cudaEventRecord(stop);
	cudaEventSynchronize(stop);
	float milliseconds = 0;
	cudaEventElapsedTime(&milliseconds, start, stop);
	printf("%f\n", milliseconds/1000.0);
}

char* read_input(char* input) {
	/* Open file and read in first 2 things*/
	FILE* input_file = fopen(input, "r");
	char* result;
	if (input_file != NULL) {
		fseek(input_file, 0L, SEEK_END);
		long s = ftell(input_file);
		rewind(input_file);

		result = (char*) malloc(s);
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

	free(state);
  	free(key);
}
