### BMUF with trimed_mean algo for Fed_learn 

import torch
import numpy as np


n_attackers = 3

block_lr = 1                   
block_momentum = 0.875         
global_sync_iter = 50          
warmup_iterations = 500
use_nbm =False
# average_sync = False
# distributed_world_size = 40 

# _num_updates = 0

# def step(self, closure=None):
#     """Performs a single optimization step."""
#     self._optimizer.step(closure)
#     self.set_num_updates(self.get_num_updates() + 1)
#     if self._is_warmup_end():
#         self._warmup_sync()
#     elif self._is_bmuf_iter():
#         self._block_sync()


#################################################### CLIENT #######################################################
######################  (0) for client compute local  --> return  local_weight(i) - global_weight(i-1)  --> flatten
# @torch.no_grad()
# def reset_local_data(params):
#     # (Step-0) Initialize global momentum parameters and store global copy on each gpu
#     global_params = [torch.zeros_like(p.data) for p in params]
#     smoothed_grads = [p.data.new_zeros(p.data.size()) for p in params]
#     grads  = [p.data.new_zeros(p.data.size()) for p in params]
    
#     # saving the global model locally for calculating gradient during bmuf sync
#     for param, global_param in zip(params, global_params):
#         global_param.copy_(param.data)

#@torch.no_grad()
# def _calc_grad(self):
#         # global_params is basically the global copy from the previously finished synchronisation. 
#         # param.data is local parameter after block_sync_freq for the local gpu. 
#         # so grad is difference between previously synced model and currrent local model.
#         for index, (param, global_param) in enumerate(
#             zip(self.params, self.global_params)
#         ):
#             self.grads[index] = global_param - param.data

######################  (1) client call "Peer.sendUpdateToAggregator( my_grade ) ( global_weight )“  send to Aggregator 
######################  (2) aggregator get enough update to call " Py_trimed_Bmuf(Global_weights,grads , num_client_sendupdate) “


################################################## AGGREGATOR ########################################################

def reshape_flatten_arg(arg):
    return self.model.reshape(arg)


def trimed_Bmuf( flaten_Global_weights , flaten_grads   ) :
    global smoothed_grads
    #######################(0) grads -- after calc_grads on local client --> turn to list(tensor)
    Global_weights = reshape_flatten_arg(np.array(flaten_Global_weights))  # [tensor]
    
    #######################(1) avg\mean Grads use tr_mean() instead of avg()
    
    flaten_grads_after_byzantine  = tr_mean( torch.tensor(flaten_grads) , n_attackers )  # tensor
    
    grads = reshape_flatten_arg(flaten_grads_after_byzantine.numpy())  # [tensor]
    
    smoothed_grads = [p.data.new_zeros(p.data.size()) for p in grads ] # create delta that shape like grads [tensor]
    
    #######################(2) Calculate global momentum and update the global model
     
    new_Global_weights = update_global_model( Global_weights, smoothed_grads , grads  )
    
    return new_Global_weights
    
        
# trimmed_mean for byzantine  tr_mean( list[ list[]]  , n)
def tr_mean(all_updates, n_attackers):
    all_updates = np.array(all_updates)
    all_updates.sort(axis=0)[0]
    out = torch.mean(all_updates[n_attackers:-n_attackers], 0) if n_attackers else torch.mean(all_updates,0)
    return out


@torch.no_grad()
def update_global_model(Global_weights , smoothed_grads , grads):
    
    New_param = [torch.zeros_like(p.data) for  p in Global_weights]
    for index, (param, global_param, smoothed_grad, grad) in enumerate(
        zip(
            New_param,
            Global_weights,
            smoothed_grads,
            grads,
       )
    ):
        # global_param is basically last syncrhornized parameter. though
        # smoothed_grad is local, all processes will have same value of
        # smoothed_grad and hence param is globally synchronized copy.
        # smoothed_grad(t) = BM * smoothed_grad(t-1) + BM_lr * grad(t)
        smoothed_grad = block_momentum * smoothed_grad + block_lr * grad
        param.data.copy_(global_param - smoothed_grad)
        
        # A Nesterov momentum here is to do a partial weight update before
        # calculating the gradient
        if use_nbm:
            param.data.copy_(param.data - block_momentum * smoothed_grad)
            
        # backup for the next synchronization.
        smoothed_grads[index] = smoothed_grad
    
    return New_param
