#!/usr/bin/env python3
"""Generate thesis benchmark figures from benchmark_results.csv.

Run from any directory:
    python3 benchmarks/generate_figures.py

Output files in docs/thesis/figures/:
  speedup_p16.png          -- overview bar chart at P=16
  bench_<name>.png (x10)   -- individual scalability curve per benchmark
"""

import csv
import math
import os
from collections import defaultdict

import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np

# ── Constants ─────────────────────────────────────────────────────────────────
T_CRIT   = 2.262        # t_{0.025, 9}
DPI      = 180
C_MANUAL = '#2C7BB6'    # steel-blue
C_GMP    = '#D7191C'    # crimson
C_GMP2   = '#F07D1B'    # orange  (second GMP variant: task_depend)
C_IDEAL  = '#AAAAAA'    # gray
PROCS    = [1, 2, 4, 8, 16]

REPO_ROOT    = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
CSV_PATH     = os.path.join(REPO_ROOT, 'benchmark_results.csv')
LOC_CSV_PATH = os.path.join(REPO_ROOT, 'loc_results.csv')
FIG_DIR      = os.path.join(REPO_ROOT, 'docs', 'thesis', 'figures')

# ── Global style ──────────────────────────────────────────────────────────────
plt.rcParams.update({
    'font.family':       'DejaVu Sans',
    'font.size':         10,
    'axes.titlesize':    12,
    'axes.titleweight':  'bold',
    'axes.labelsize':    10,
    'axes.spines.top':   False,
    'axes.spines.right': False,
    'legend.fontsize':   9,
    'legend.framealpha': 0.92,
    'xtick.labelsize':   9,
    'ytick.labelsize':   9,
    'lines.linewidth':   2.0,
    'lines.markersize':  6,
    'grid.alpha':        0.35,
    'grid.linestyle':    '--',
    'figure.dpi':        DPI,
})

# ── Data loading ──────────────────────────────────────────────────────────────
raw = defaultdict(list)
with open(CSV_PATH, newline='') as f:
    reader = csv.DictReader(f)
    for row in reader:
        key = (row['benchmark'], int(row['procs']), row['variant'])
        raw[key].append(int(row['time_ns']) / 1e6)


def _stats(values):
    n = len(values)
    mean = sum(values) / n
    var  = sum((x - mean) ** 2 for x in values) / (n - 1) if n > 1 else 0.0
    cv   = math.sqrt(var) / mean if mean else 0.0
    return mean, cv


_agg = {key: _stats(times) for key, times in raw.items()}


def speedup(bench, procs, variant):
    """Return (speedup, CI95) vs same-P sequential baseline, or (nan, nan)."""
    seq = _agg.get((bench, procs, 'seq'))
    par = _agg.get((bench, procs, variant))
    if seq is None or par is None or par[0] == 0:
        return float('nan'), float('nan')
    seq_m, seq_cv = seq
    par_m, par_cv = par
    sp = seq_m / par_m
    ci = T_CRIT * sp * math.sqrt(seq_cv**2 + par_cv**2)
    return sp, ci


def sp_curve(bench, variant):
    return np.array([speedup(bench, p, variant)[0] for p in PROCS], dtype=float)


# ── Figure: speedup_p16.png ───────────────────────────────────────────────────
def fig_speedup_p16():
    # Order: wins -> parity -> loses
    items = [
        ('reduce',     'gompher',     'Reduce'),
        ('nqueens',    'gompher',     'N-Queens'),
        ('matmul',     'gompher',     'MatMul'),
        ('pipeline',   'gompher',     'Pipeline'),
        ('mergesort',  'gompher',     'MergeSort'),
        ('sections',   'gompher',     'Sections'),
        ('montecarlo', 'gompher',     'MonteCarlo'),
        ('quicksort',  'gompher',     'QuickSort'),
        ('prefixsum',  'gompher',     'PrefixSum'),
        ('heavyreduce',  'taskloop',    'HeavyReduce (taskloop)'),
        ('heavyreduce',  'task_depend', 'HeavyReduce (depend)'),
    ]

    labels = [it[2] for it in items]
    n = len(labels)
    s_man  = []; ci_man = []
    s_gmp  = []; ci_gmp = []
    for bench, gmp_var, _ in items:
        sm, cm = speedup(bench, 16, 'manual')
        sg, cg = speedup(bench, 16, gmp_var)
        s_man.append(sm);  ci_man.append(cm)
        s_gmp.append(sg);  ci_gmp.append(cg)

    s_man  = np.array(s_man,  dtype=float)
    ci_man = np.array(ci_man, dtype=float)
    s_gmp  = np.array(s_gmp,  dtype=float)
    ci_gmp = np.array(ci_gmp, dtype=float)

    y  = np.arange(n)
    bh = 0.35
    fig, ax = plt.subplots(figsize=(10, 7))

    ax.barh(y + bh / 2, s_man, bh, xerr=ci_man, color=C_MANUAL,
            label='Manual', capsize=3,
            error_kw={'elinewidth': 1.3, 'ecolor': '#1a4f7a'})
    ax.barh(y - bh / 2, s_gmp, bh, xerr=ci_gmp, color=C_GMP,
            label='GompherMP', capsize=3,
            error_kw={'elinewidth': 1.3, 'ecolor': '#8b0000'})

    ax.axvline(1.0, color='k', linestyle='--', linewidth=0.9, alpha=0.5)
    ax.axhline(3.5, color='#cccccc', linewidth=0.9, zorder=0)
    ax.axhline(5.5, color='#cccccc', linewidth=0.9, zorder=0)

    xmax = float(np.nanmax(np.concatenate([s_man + ci_man, s_gmp + ci_gmp]))) * 1.04
    ax.text(xmax * 0.99, 1.5,  'GompherMP vence',            ha='right', va='center',
            fontsize=8, color='#444', fontstyle='italic')
    ax.text(xmax * 0.99, 4.5,  'Paridad estadistica',         ha='right', va='center',
            fontsize=8, color='#444', fontstyle='italic')
    ax.text(xmax * 0.99, 8.0,  'GompherMP cede rendimiento',  ha='right', va='center',
            fontsize=8, color='#444', fontstyle='italic')

    ax.set_yticks(y)
    ax.set_yticklabels(labels)
    ax.set_xlabel('Speedup ($S_{16}$)')
    ax.set_title('Comparacion de speedup a $P = 16$ procesadores (IC 95%)')
    ax.legend(loc='lower right')
    ax.grid(axis='x')
    ax.set_xlim(0, xmax)

    fig.tight_layout()
    _save(fig, 'speedup_p16.png')


# ── Per-benchmark scalability figures ─────────────────────────────────────────
# Each entry: (bench_key, display_title, [(variant, label, color, marker), ...])
BENCH_CONFIGS = [
    ('matmul',    'MatMul',    [
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
    ('prefixsum', 'PrefixSum', [
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
    ('mergesort', 'MergeSort', [
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
    ('heavyreduce', 'HeavyReduce', [
        ('manual',      'Manual',               C_MANUAL, '-o'),
        ('taskloop',    'GompherMP (taskloop)',  C_GMP,    '-s'),
        ('task_depend', 'GompherMP (depend)',    C_GMP2,   '-^'),
    ]),
    ('sections',  'Sections',  [
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
    ('quicksort', 'QuickSort', [
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
    ('pipeline',  'Pipeline',  [
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
    ('nqueens',   'N-Queens',  [
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
    ('montecarlo','MonteCarlo',[
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
    ('reduce',    'Reduce',    [
        ('manual',   'Manual',    C_MANUAL, '-o'),
        ('gompher',  'GompherMP', C_GMP,    '-s'),
    ]),
]


def fig_bench(bench, title, variants):
    xs = np.array(PROCS, dtype=float)
    fig, ax = plt.subplots(figsize=(7, 4.5))

    ax.plot(xs, xs, color=C_IDEAL, linestyle='--', linewidth=1.2,
            label='Escalado lineal', zorder=1)

    for var, label, color, marker in variants:
        ys = sp_curve(bench, var)
        ax.plot(xs, ys, marker, color=color, label=label, zorder=3)

    ax.set_title(title)
    ax.set_xlabel('Procesadores ($P$)')
    ax.set_ylabel('Speedup ($S_P$)')
    ax.set_xticks(PROCS)
    ax.legend(fontsize=9)
    ax.grid(True)
    ax.set_xlim(0.5, 17)
    ax.set_ylim(bottom=0)

    fig.tight_layout()
    _save(fig, f'bench_{bench}.png')


# ── Figure: loc_comparison.png ───────────────────────────────────────────────
def fig_loc_comparison():
    # Ordered: most reduction (bottom) → most increase (top), as in loc_results.csv
    data = []
    with open(LOC_CSV_PATH, newline='') as f:
        for row in csv.DictReader(f):
            data.append((row['benchmark'], int(row['loc_manual']), int(row['loc_gompher'])))

    labels  = [d[0] for d in data]
    loc_man = np.array([d[1] for d in data], dtype=float)
    loc_gmp = np.array([d[2] for d in data], dtype=float)

    n  = len(labels)
    y  = np.arange(n)
    bh = 0.35

    fig, ax = plt.subplots(figsize=(9, 7))

    ax.barh(y + bh / 2, loc_man, bh, color=C_MANUAL, label='Manual')
    ax.barh(y - bh / 2, loc_gmp, bh, color=C_GMP,    label='GompherMP')

    xmax = float(max(max(loc_man), max(loc_gmp))) * 1.45

    # delta-LoC annotations
    for i, (lm, lg) in enumerate(zip(loc_man, loc_gmp)):
        delta = (lg - lm) / lm * 100
        sign  = '+' if delta > 0 else ''
        txt   = f'{sign}{delta:.0f}%'
        col   = '#8b0000' if delta > 0 else '#1a6b1a'
        ax.text(max(lm, lg) + 1.2, i, txt,
                va='center', fontsize=9, color=col, fontweight='bold')

    # category separators and labels
    ax.axhline(3.5, color='#cccccc', linewidth=0.9, zorder=0)
    ax.axhline(6.5, color='#cccccc', linewidth=0.9, zorder=0)
    ax.text(xmax * 0.99, 1.5, 'GompherMP reduce LoC',      ha='right',
            va='center', fontsize=8, color='#444', fontstyle='italic')
    ax.text(xmax * 0.99, 5.0, 'Paridad',                   ha='right',
            va='center', fontsize=8, color='#444', fontstyle='italic')
    ax.text(xmax * 0.99, 8.5, 'GompherMP incrementa LoC',  ha='right',
            va='center', fontsize=8, color='#444', fontstyle='italic')

    ax.set_yticks(y)
    ax.set_yticklabels(labels)
    ax.set_xlabel('Líneas de código (LoC)')
    ax.set_title('Comparación de LoC: paralelo manual vs. GompherMP')
    ax.legend(loc='lower right')
    ax.grid(axis='x')
    ax.set_xlim(0, xmax)

    fig.tight_layout()
    _save(fig, 'loc_comparison.png')


# ── Helpers ───────────────────────────────────────────────────────────────────
def _save(fig, name):
    path = os.path.join(FIG_DIR, name)
    fig.savefig(path, dpi=DPI, bbox_inches='tight')
    plt.close(fig)
    print(f'  ok {name}')


# ── Main ──────────────────────────────────────────────────────────────────────
if __name__ == '__main__':
    print('Generating thesis figures ...')
    fig_speedup_p16()
    for bench, title, variants in BENCH_CONFIGS:
        fig_bench(bench, title, variants)
    fig_loc_comparison()
    print('Done.')
