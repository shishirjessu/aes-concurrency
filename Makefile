all: aes_seq.c
	gcc -g -Wall -o aes_seq aes_seq.c

clean:
	$(RM) aes_seq
run:
	./aes_seq
