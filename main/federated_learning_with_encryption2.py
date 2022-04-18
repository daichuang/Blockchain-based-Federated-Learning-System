
import numpy as np
from sklearn.datasets import load_diabetes

import phe as paillier
import sys
seed = 43
np.random.seed(seed)
import torch
dtype = torch.cuda.DoubleTensor if torch.cuda.is_available() else torch.FloatTensor

def get_data(n_clients):
    """
    Import the dataset via sklearn, shuffle and split train/test.
    Return training, target lists for `n_clients` and a holdout test set
    """
    print("Loading data")
    diabetes = load_diabetes()
    y = diabetes.target
    X = diabetes.data
    # Add constant to emulate intercept
    X = np.c_[X, np.ones(X.shape[0])]

    # The features are already preprocessed
    # Shuffle
    perm = np.random.permutation(X.shape[0])
    X, y = X[perm, :], y[perm]

    # Select test at random
    test_size = 150
    test_idx = np.random.choice(X.shape[0], size=test_size, replace=False)
    train_idx = np.ones(X.shape[0], dtype=bool)
    train_idx[test_idx] = False
    X_test, y_test = X[test_idx, :], y[test_idx]
    X_train, y_train = X[train_idx, :], y[train_idx]

    # Split train among multiple clients.
    # The selection is not at random. We simulate the fact that each client
    # sees a potentially very different sample of patients.
    X, y = [], []
    step = int(X_train.shape[0] / n_clients)
    for c in range(n_clients):
        X.append(X_train[step * c: step * (c + 1), :])
        y.append(y_train[step * c: step * (c + 1)])

    return X, y, X_test, y_test


def mean_square_error(y_pred, y):
    """ 1/m * \sum_{i=1..m} (y_pred_i - y_i)^2 """
    return np.mean((y - y_pred) ** 2)


def encrypt_vector(public_key, x):
    return [public_key.encrypt(i) for i in x]


def decrypt_vector(private_key, x):
    return np.array([private_key.decrypt(i) for i in x])


def sum_encrypted_vectors(x, y):
    if len(x) != len(y):
        raise ValueError('Encrypted vectors must have the same size')
    return [x[i] + y[i] for i in range(len(x))]


class Server:
    """Private key holder. Decrypts the average gradient"""

    def __init__(self, key_length):
         keypair = paillier.generate_paillier_keypair(n_length=key_length)
         self.pubkey, self.privkey = keypair

    def decrypt_aggregate(self, input_model, n_clients):
        return decrypt_vector(self.privkey, input_model) / n_clients


class Client:
    """Runs linear regression with local data or by gradient steps,
    where gradient can be passed in.

    Using public key can encrypt locally computed gradients.
    """

    def __init__(self, name, X, y, pubkey):
        self.name = name
        self.pubkey = pubkey
        self.X, self.y = X, y
        self.weights = np.zeros(X.shape[1])

    def fit(self, n_iter, eta=0.01, pre = [],y_test=[]):
        """Linear regression for n_iter"""
        out = []
        for _ in range(n_iter):
            gradient = self.compute_gradient()
            self.gradient_step(gradient, eta)
            if len(pre):
                y_pred = self.predict(pre)
                mse = mean_square_error(y_pred, y_test)
                out.append(mse)
        return out
            
    def gradient_step(self, gradient, eta=0.01):
        """Update the model with the given gradient"""
        self.weights -= eta * gradient

    def compute_gradient(self):
        """Compute the gradient of the current model using the training set
        """
        delta = self.predict(self.X) - self.y
        return delta.dot(self.X) / len(self.X)

    def predict(self, X):
        """Score test data"""
        return X.dot(self.weights)

    def encrypted_gradient(self, sum_to=None):
        """Compute and encrypt gradient.

        When `sum_to` is given, sum the encrypted gradient to it, assumed
        to be another vector of the same size
        """
        gradient = self.compute_gradient()
        gradient = torch.as_tensor(gradient).type(dtype)
        encrypted_gradient = encrypt_vector(self.pubkey, gradient)
         

        if sum_to is not None:
            return sum_encrypted_vectors(sum_to, encrypted_gradient)
        else:
            return encrypted_gradient

import time

def federated_learning(X, y, X_test, y_test, config,timeit = 0):
    n_clients = config['n_clients']
    n_iter = config['n_iter']
    names = ['Hospital {}'.format(i) for i in range(1, n_clients + 1)]

    # Instantiate the server and generate private and public keys
    # NOTE: using smaller keys sizes wouldn't be cryptographically safe
    server = Server(key_length=config['key_length'])

    # Instantiate the clients.
    # Each client gets the public key at creation and its own local dataset
    clients = []
    for i in range(n_clients):
        clients.append(Client(names[i], X[i], y[i], server.pubkey))

    # The federated learning with gradient descent
    print('Running distributed gradient aggregation for {:d} iterations'
          .format(n_iter))
    cost_encrypt_aggr=0
    cost_encrypted_gradient=0
    cost_decrypt_aggregate=0
    allout = []
    for i in range(n_iter):

        # Compute gradients, encrypt and aggregate
        time_start=time.time()
        encrypt_aggr = clients[0].encrypted_gradient(sum_to=None)
        cost_encrypt_aggr+=time.time()-time_start
        #print(encrypt_aggr)
        #exit(0)
        
        time_start=time.time()
        for c in clients[1:]:
            encrypt_aggr = c.encrypted_gradient(sum_to=encrypt_aggr)
        cost_encrypted_gradient+=time.time()-time_start
 
    
        time_start=time.time()
        # Send aggregate to server and decrypt it
        aggr = server.decrypt_aggregate(encrypt_aggr, n_clients)
        #print(aggr)
        #sys.exit(0)
        cost_decrypt_aggregate+=time.time()-time_start
        #print(len(encrypt_aggr))
        #sys.exit(0)
        # Take gradient steps
        
        for c in clients:
            c.gradient_step(aggr, config['eta'])
        print(i)
        if not timeit:
            out = []
            for c in clients:
                y_pred = c.predict(X_test)
                mse = mean_square_error(y_pred, y_test)
                out.append(mse)
            allout.append(out)
    alltime = [cost_encrypt_aggr,cost_encrypted_gradient,cost_decrypt_aggregate]
    print('Error (MSE) that each client gets after running the protocol:')
    for c in clients:
        y_pred = c.predict(X_test)
        mse = mean_square_error(y_pred, y_test)
        print('{:s}:\t{:.2f}'.format(c.name, mse))
    return allout,alltime

def local_learning(X, y, X_test, y_test, config):
    n_clients = config['n_clients']
    names = ['Hospital {}'.format(i) for i in range(1, n_clients + 1)]

    # Instantiate the clients.
    # Each client gets the public key at creation and its own local dataset
    clients = []
    for i in range(n_clients):
        clients.append(Client(names[i], X[i], y[i], None))

    # Each client trains a linear regressor on its own data
    print('Error (MSE) that each client gets on test set by '
          'training only on own local data:')
    allout = []
    for c in clients:
        out = c.fit(config['n_iter'], config['eta'],pre = X_test, y_test = y_test)
        allout.append(out)
        y_pred = c.predict(X_test)
        mse = mean_square_error(y_pred, y_test)
        print('{:s}:\t{:.2f}'.format(c.name, mse))
    return allout
    
if __name__ == '__main__':
    config = {
        'n_clients': 5,
        'key_length': 1024,
        'n_iter': 800,
        'eta': 0.5,
    }
    # load data, train/test split and split training data between clients
    X, y, X_test, y_test = get_data(n_clients=config['n_clients'])
    # first each hospital learns a model on its respective dataset for comparison.
    allout = local_learning(X, y, X_test, y_test, config)
    import pandas as pd
    data = pd.DataFrame(allout).T
    
    #import profile

    alldata,times = federated_learning(X, y, X_test, y_test, config)
    all_data = pd.DataFrame(alldata)
    data.to_csv("central5.csv")
    all_data.to_csv("federal5.csv")
    print(times)
    
    
    
    
    #profile.run('federated_learning(X, y, X_test, y_test, config)',"federated_learning")
    # and now the full glory of federated learning
    #federated_learning(X, y, X_test, y_test, config)
