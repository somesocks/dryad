# diagnostics-11--every-x-per-key

Ensures diagnostics `every_x_per_key` behaves correctly on `os.link`, succeeding with a large per-key period and failing when the periodic hit lands on repeated heap-file links.
