all: aes_seq.c
	gcc -g -Wall -o aes_seq aes_seq.c

clean:
	$(RM) aes_seq
run:
	./aes_seq

run_go_seq:
	go run useful_stuff.go aes_seq.go

run_go_par:
	go run useful_stuff.go aes_seq.go
