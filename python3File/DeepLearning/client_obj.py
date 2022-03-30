from __future__ import division
import numpy as np
import client
import pdb
import emcee
from softmax_model import SoftmaxModel
from mnist_dnn import mnist_dnn
import datasets
from torchsummary import summary
import torch.nn as nn
import random
from outliers import smirnov_grubbs as grubbs
import numpy as np
import torch


def CreateClient(dataset, filename, batch_size , epoch , disable_dp , epsilon):
    global myclient
    # D_in = datasets.get_num_features(dataset)  #784
    # D_out = datasets.get_num_classes(dataset)  #10
    # model = SoftmaxModel(D_in, D_out)
    
    model = mnist_dnn()
    # print(summary(model ,(28,28)))
    myclient = client.Client(dataset, filename, batch_size, model,  epoch , disable_dp , epsilon )

def getNumFeature(dataset):
    # nParams = 7850
    nParams = datasets.get_num_params(dataset)
    return nParams


# local param - w(i-1)
def GetLocalParam(ww):
    global myclient
    weights = np.array(ww)
    # print("ACc before com",getTestAcc(ww))
    myclient.updateModel(weights)
    newParam = myclient.getNewParam()
    return (newParam - ww) 
    #* len(myclient.trainloader.dataset)

def GetLocalDataNum():
    return len(myclient.trainloader.dataset)


def GetLocalTestDataNum():
    return len(myclient.testloader.dataset)

def getInitWeight():
    myclient.model.apply(weights_init)
    layers = np.zeros(0)
    for _, param in myclient.model.named_parameters():
        if param.requires_grad:
            layers = np.concatenate((layers, param.data.numpy().flatten()), axis=None)
    return layers


def setup_seed(seed):
     torch.manual_seed(seed)
     torch.cuda.manual_seed_all(seed)
     torch.backends.cudnn.deterministic = True


def weights_init(m):
    setup_seed(random.random())
    classname = m.__class__.__name__
    if classname.find('Conv2d') != -1:
        nn.init.xavier_normal_(m.weight.data)
        nn.init.constant_(m.bias.data, 0.0)
    elif classname.find('Linear') != -1:
        nn.init.xavier_normal_(m.weight)
        nn.init.constant_(m.bias, 0.0)

#Secure aggregation
def SecureAggregate(flaten_Global_weights , flaten_grads ):
    trimed_Bmuf( flaten_Global_weights , flaten_grads , myclient )


# Acc
def getTestErr(ww):
    global myclient
    weights = np.array(ww)
    myclient.updateModel(weights)
    return myclient.getTestErr()

def getTestAcc(ww):
    global myclient
    weights = np.array(ww)
    myclient.updateModel(weights)
    return myclient.getTestAcc()

# Secure methods
#alpha = 0.05 before
# THIS METHOD NOT WORK 
def outlierDetect(Goarray) :
    data = np.array(Goarray)
    list_index = grubbs.two_sided_test_indices(data, alpha=0.05)
    # for value in list_index:
    #     list_index[list_index.index(value)] = double(value)
    
    return list_index

def smoothedGrag_detect(similarity_array,n_byzantine):
    G = []
    for i in range(len(similarity_array)):
        G.append(i) 
    R = [1]
    B = []
    ksi = 2
    ksi_delta = 0.5
    while len(R) != 0 :
        R = []
        new_array = []
        if len(G) == 2 * n_byzantine + 1 :
            break
        for i in G :
            new_array.append(similarity_array[i])
        mean = np.mean(new_array)
        median = np.median(new_array)
        std = np.std(new_array ,  ddof=1)
        if mean < median :
            for k in G :
                if similarity_array[k] < median - ksi * std :
                    if len(G) > 2 * n_byzantine + 1 :
                        R.append(k)
                        G.remove(k)
        else :
            for k in G :
                if similarity_array[k] > median + ksi * std :
                    if len(G) > 2 * n_byzantine + 1 :
                        R.append(k)
                        G.remove(k)
                    
        ksi += ksi_delta
        B[1:1] = R
    
    return B
    
def resetVar():
    global i 
    i = 0
    
def gradCollector(accept_grad , data_num , num_update):
    global accept_grad_list
    global data_num_list 
    global i 
    if i < num_update :
        if i == 0 :
            accept_grad_list = np.array(accept_grad)
            data_num_list = [data_num]
            i += 1 
        else :
            accept_grad_list = np.vstack((accept_grad_list , np.array(accept_grad)))
            data_num_list.append(data_num)
            i += 1
    # return len(accept_grad_list)
            
def gradTrim(num_update ,n_byzantine , num_feature):
    accept_grad_tensor = torch.from_numpy(accept_grad_list)
    param_med = torch.median(accept_grad_tensor , dim = 0 )[0]
    
    sort_idx = torch.argsort(torch.abs(accept_grad_tensor - param_med) , dim=0)
    sorted_params = accept_grad_list[sort_idx , torch.arange(num_feature)[None ,:]]

    sorted_params_after_trim = sorted_params[:num_update - 2 * n_byzantine]
    sort_idx_num = sort_idx[:num_update - 2 * n_byzantine].numpy()

    for x in np.nditer(sort_idx_num, op_flags=['readwrite']):
        x[...] = data_num_list[x]

    weight_param = sort_idx_num * sorted_params_after_trim
    res = np.sum(weight_param,0)
        
    return res

if __name__ == '__main__':
    
    b = smoothedGrag_detect( [-0.0508339580406963,0.06702821114240805,0.016717913582610915,0.04541585586939311,0.03362754327730253,0.03440675636482688,0.055077516722803074,0.030845490862268433,0.013392710703108312,0.0441366936538565,0.017245009919088162,0.043686708415391885,0.06465272619396503,-0.06556358844612668,-0.06556358844612668,-0.06556358844612668,-0.06556358844612668,-0.06556358844612668,-0.06556358844612668],6)
    print(b)
    # batch_size = 50
    
    # CreateClient("mnist", "mnist0", batch_size , 3 , False , 1 )
    
    # myclient.model.apply(weights_init)
    
    # layers = np.zeros(0)
    # for name, param in myclient.model.named_parameters():
    #     if param.requires_grad:
    #         layers = np.concatenate((layers, param.data.numpy().flatten()), axis=None)
    # print(layers)
    # print(outlierDetect([0.028112280504984512,0.03152145105885309 ,0.0242043589406434 ,0.05541787433396907, 0.0412890148968361, 0.05768263845239046, 0.047364014567759176 ,0.07146752080093688 ,0.061519107163882425, 0.054447943923542776 ,0.045638633281139124 ,0.024712587598947143, 0.06308616943651169 ,0.06033066856270841, 0.0392993564270343 ,0.04693398960480092, 0.03171047997227712 ,-0.030122538630654987 ,-0.030122538630654987]))

    # print(GetLocalParam(layers))
    # print(GetLocalDataNum())
    # print(GetLocalTestDataNum())