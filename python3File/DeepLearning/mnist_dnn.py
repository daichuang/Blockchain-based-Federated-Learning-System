
import torch
import torch.nn as nn
import torch.nn.functional as F
import numpy as np

class mnist_dnn(nn.Module):
  
  def __init__(self):
    super(mnist_dnn, self).__init__()

    self.linear1 = nn.Linear(784, 32)
    self.linear2 = nn.Linear(32, 10)

  def forward(self, x):
    x = np.reshape(x, (x.shape[0], 784))
    x = self.linear1(x)
    x = F.relu(x)
    out = self.linear2(x)
    return out

  def reshape(self, flat_gradient):
        layers = []
        l1 = 784*32
        l1_b = 32
        l2 = 32*10
        l2_b = 10
        
        layers.append( torch.from_numpy(np.reshape(flat_gradient[0:l1], (32 , 784 ))).type(torch.FloatTensor) )
        layers.append( torch.from_numpy(flat_gradient[l1 : l1+l1_b]).type(torch.FloatTensor) ) 
        layers.append( torch.from_numpy(np.reshape(flat_gradient[l1+l1_b:l1+l1_b+l2], (10 ,32 ))).type(torch.FloatTensor) )
        layers.append( torch.from_numpy(flat_gradient[l1+l1_b+l2 : l1+l1_b+l2+l2_b]).type(torch.FloatTensor) )
   
        return layers