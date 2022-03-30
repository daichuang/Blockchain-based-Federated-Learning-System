from mnist_dataset import MNISTDataset

def get_dataset(dataset):
    if dataset == "mnist":
        return MNISTDataset
    else:
        print("Error: dataset " + dataset + "not defined")
    
def get_num_params(dataset):
    if dataset == "mnist":
        return 25450
    else:
        print("Error: dataset " + dataset + "not defined")

def get_num_features(dataset):
    if dataset == "mnist":
        return 784
    else:
        print("Error: dataset " + dataset + "not defined")
    
def get_num_classes(dataset):
    if dataset == "mnist":
        return 10
    else: 
        print("Error: dataset " + dataset + "not defined")