from phe import paillier

# generate a public key and private key pair
public_key, private_key = paillier.generate_paillier_keypair()

# start encrypting numbers
secret_number_list = [3.141592653, 300, -4.6e-12]
encrypted_number_list = [public_key.encrypt(x) for x in secret_number_list]
print(encrypted_number_list)
# decrypt encrypted number and print
print([private_key.decrypt(x) for x in encrypted_number_list])

