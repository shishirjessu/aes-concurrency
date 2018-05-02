keypath = input/key.txt
smallpath = input/small.txt
medpath = input/med.txt
lgpath = input/lg.txt
workers = 1

all: aes_seq.c
	gcc -g -Wall -o aes_seq aes_seq.c

clean:
	$(RM) aes_seq

run:
	./aes_seq

run_go_seq_sm:
	go run useful_stuff.go aes_seq.go $(keypath) $(smallpath)

run_go_par_sm:
	go run useful_stuff.go aes_par.go $(keypath) $(smallpath) $(workers)

run_go_seq_md:
	go run useful_stuff.go aes_seq.go $(keypath) $(medpath)

run_go_par_md:
	go run useful_stuff.go aes_par.go $(keypath) $(medpath) $(workers)

run_go_seq_lg:
	go run useful_stuff.go aes_seq.go $(keypath) $(lgpath)

run_go_par_lg:
	go run useful_stuff.go aes_par.go $(keypath) $(lgpath) $(workers)
