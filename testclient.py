#!/usr/bin/python3

import json
import socket
import ssl

# CREATE COMMAND
cmd = {
  "op": "is_authorized",
  "email": "ameyabhurke@outlook.com",
  "password": "password",
  "resource": "/home/abhurke/userd"
}
cmdStr = json.dumps(cmd)

host_addr = '127.0.0.1'
host_port = 9669
server_sni_hostname = 'openspock.org'
server_cert = '/home/abhurke/userd/server.crt'
client_cert = 'client.crt'
client_key = 'client.key'

context = ssl.create_default_context(ssl.Purpose.SERVER_AUTH, cafile=server_cert)
#context.load_cert_chain(certfile=client_cert, keyfile=client_key)

s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
conn = context.wrap_socket(s, server_side=False, server_hostname=server_sni_hostname)
conn.connect((host_addr, host_port))
print("SSL established. Peer: {}".format(conn.getpeercert()))
print("sending ...")
print(cmdStr)
conn.send(str.encode(cmdStr))
response = conn.recv(1024)
print("response: " + response.decode())
print("Closing connection")
conn.close()