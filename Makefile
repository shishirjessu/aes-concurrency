all: aes_seq.c
	gcc -g -Wall -o aes_seq aes_seq.c

clean:
	$(RM) aes_seq
run:
	./aes_seq

run_go_seq_sm:
	go run useful_stuff.go aes_seq.go input/key.txt input/small.txt

run_go_par_sm:
	go run useful_stuff.go aes_par.go input/key.txt input/small.txt $(workers)
