import sys
import collections

def summarize_lcov(filepath):
    files = {}
    current_file = None
    lines_found = 0
    lines_hit = 0

    with open(filepath, 'r') as f:
        for line in f:
            if line.startswith('SF:'):
                current_file = line.strip().split(':')[1]
                files[current_file] = {'LF': 0, 'LH': 0}
            elif line.startswith('LF:'):
                lf = int(line.strip().split(':')[1])
                files[current_file]['LF'] = lf
                lines_found += lf
            elif line.startswith('LH:'):
                lh = int(line.strip().split(':')[1])
                files[current_file]['LH'] = lh
                lines_hit += lh

    print(f"Total Coverage: {(lines_hit/lines_found)*100:.2f}% ({lines_hit}/{lines_found})")
    for file, cov in sorted(files.items(), key=lambda x: x[1]['LH']/x[1]['LF'] if x[1]['LF'] else 0):
        if cov['LF'] > 0:
            percent = (cov['LH'] / cov['LF']) * 100
            if percent < 95.0:
                print(f"Under 95% - {percent:6.2f}% ({cov['LH']}/{cov['LF']}) - {file}")

if __name__ == '__main__':
    summarize_lcov(sys.argv[1])
