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
    print(size, i)
    temp = []
    for _ in range(10):
        temp.append(float(check_output((rules[0] + types[0] + size), shell=True).decode().split('\n')[1]))
    seq_times_c.append(sum(temp)/10)

    temp = []
    for _ in range(10):
        temp.append(float(check_output((rules[0] + types[1] + size), shell=True).decode().split('\n')[1]))
    par_times_c.append(sum(temp)/10)

    temp = []
    for _ in range(10):
        temp.append(float(check_output((rules[1] + types[0] + size), shell=True).decode().split('\n')[1]))
    seq_times_go.append(sum(temp)/10)

    temp = []
    for _ in range(10):
        temp.append(float(check_output((rules[1] + types[1] + size), shell=True).decode().split('\n')[1]))
    par_times_go.append(sum(temp)/10)



par_speedup_c = [seq_times_c[i]/par_times_c[i] for i in range(len(par_times_c))]
par_speedup_go = [seq_times_go[i]/par_times_go[i] for i in range(len(par_times_go))]

f = open('output.txt', 'w+')

f.write('seq times c\n')
f.write(str(seq_times_c))
f.write('\n')

f.write('seq times go\n')
f.write(str(seq_times_go))
f.write('\n')

f.write('par times c\n')
f.write(str(par_times_c))
f.write('\n')

f.write('par times go\n')
f.write(str(par_times_go))
f.write('\n')

f.write('par speedup c\n')
f.write(str(par_speedup_c))
f.write('\n')

f.write('par speedup go\n')
f.write(str(par_speedup_go))
f.write('\n')
