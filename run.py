from subprocess import check_output

rules = ['make run_c_', 'make run_go_']
types = ['seq_', 'par_']
sizes = ['sm', 'md', 'lg', 'xlg', 'xxlg', 'xxxlg']

check_output(['make', 'clean'])
check_output(['make', 'all'])

seq_times_c = []
seq_times_go = []

par_times_c = []
par_times_go = []

for i in range(len(sizes)):
    size = sizes[i]
    temp = []
    for _ in range(10):
        temp.append(float(check_output((rules[0] + types[0] + size), shell=True).decode().split('\n')[0]).strip())
    seq_times_c.append(sum(temp)/10)

    temp = []
    for _ in range(10):
        temp.append(float(check_output((rules[0] + types[1] + size), shell=True).decode().split('\n')[0]).strip())
    par_times_c.append(sum(temp)/10)

    temp = []
    for _ in range(10):
        temp.append(float(check_output((rules[1] + types[0] + size), shell=True).decode().split('\n')[0]).strip())
    seq_times_go.append(sum(temp)/10)

    temp = []
    for _ in range(10):
        temp.append(float(check_output((rules[1] + types[1] + size), shell=True).decode().split('\n')[0]).strip())
    par_times_go.append(sum(temp)/10)


c_baseline = seq_times_c[0]
go_baseline = seq_times_go[0]

seq_speedup_c = [c_baseline/thing for thing in seq_times_c]
seq_speedup_go = [go_baseline/thing for thing in seq_times_go]

par_speedup_c = [c_baseline/thing for thing in par_times_c]
par_speedup_go = [go_baseline/thing for thing in par_times_go]

f = open('output.txt', 'w+')

f.write('seq times c\n')
f.write(str(seq_times_c))

f.write('seq times go\n')
f.write(str(seq_times_go))

f.write('par times c\n')
f.write(str(par_times_c))

f.write('par times go\n')
f.write(str(par_times_go))

f.write('seq speedup c\n')
f.write(str(seq_speedup_c))

f.write('seq speedup go\n')
f.write(str(seq_speedup_go))

f.write('par speedup c\n')
f.write(str(par_speedup_c))

f.write('par speedup go\n')
f.write(str(par_speedup_go))
