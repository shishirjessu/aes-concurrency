keypath = input/key.txt
smallpath = input/small.txt
medpath = input/med.txt
lgpath = input/lg.txt
xlgpath = input/xlg.txt
xxlgpath = input/xxlg.txt
xxxlgpath = input/xxxl.txt

all: aes_seq.c
	gcc -g -Wall -o aes_seq aes_seq.c
	nvcc -o aes-ctr aes-ctr.cu

clean:
	$(RM) aes_seq

run_c_seq_sm:
	./aes_seq $(keypath) $(smallpath)

run_c_seq_md:
	./aes_seq $(keypath) $(medpath)

run_c_seq_lg:
	./aes_seq $(keypath) $(lgpath)

run_c_par_sm:
	./aes-ctr $(keypath) $(smallpath)

run_c_par_md:
	./aes-ctr $(keypath) $(medpath)

run_c_par_lg:
	./aes-ctr $(keypath) $(lgpath)

run_go_seq_sm:
	go run useful_stuff.go aes_seq.go $(keypath) $(smallpath)

run_go_par_sm:
	go run useful_stuff.go aes_par.go $(keypath) $(smallpath)

run_go_seq_md:
	go run useful_stuff.go aes_seq.go $(keypath) $(medpath)

run_go_par_md:
	go run useful_stuff.go aes_par.go $(keypath) $(medpath)

run_go_seq_lg:
	go run useful_stuff.go aes_seq.go $(keypath) $(lgpath)

run_go_par_lg:
	go run useful_stuff.go aes_par.go $(keypath) $(lgpath)

run_go_seq_xlg:
	go run useful_stuff.go aes_seq.go $(keypath) $(xlgpath)

run_go_par_xlg:
	go run useful_stuff.go aes_par.go $(keypath) $(xlgpath)

run_go_seq_xxlg:
	go run useful_stuff.go aes_seq.go $(keypath) $(xxlgpath)

run_go_par_xxlg:
	go run useful_stuff.go aes_par.go $(keypath) $(xxlgpath)

run_go_seq_xxxlg:
	go run useful_stuff.go aes_seq.go $(keypath) $(xxxlgpath)

run_go_par_xxxlg:
	go run useful_stuff.go aes_par.go $(keypath) $(xxxlgpath)
