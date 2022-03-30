
from outliers import smirnov_grubbs as grubbs
import numpy as np

data = data = np.array([1, 8, 9, 10, 9,100])
a = grubbs.two_sided_test_indices(data, alpha=0.05)


print(np.array(a)+1)