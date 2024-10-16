from pwn import *
import threading
import re

PACKET_IN = 0
PACKET_OUT = 1
PACKET_BROADCAST = 2
PACKET_NEW_CONNECTION = 3

ANSI_WHITE = "\033[0;37m"
ANSI_CYAN =   "\033[0;36m"
ANSI_YELLOW = "\033[0;33m"

user_id = 0

context.arch = 'x86_64'

def get_incoming_messages(r: tube):
    global user_id
    info("Waiting for packages...")
    
    while True:
        while b := r.recv(16, timeout=0.2):
            t = u64(b[:8])
            l = u64(b[8:16])
            body = r.recv(l, timeout=0.5)

            if t == PACKET_OUT:
                sender_id = u64(body[:8])
                print(ANSI_CYAN + f"{sender_id}> " + ANSI_WHITE + body[8:].decode())
            elif t == PACKET_NEW_CONNECTION:
                user_id = u64(body[:8])
                warn(f"ID: {user_id}")

            else:
                print("received bad formatted packet!")

    info("Done!")

def main():
    global user_id
    r = remote("127.0.0.1", 8080)

    t = threading.Thread(target=get_incoming_messages, args=(r,))
    t.start()

    while not user_id: continue

    running = True
    while (l := input(ANSI_YELLOW + f"{user_id}> " +ANSI_WHITE)) != "\n" and running:
        try:
            target_id = int(re.findall("@[0-9]+", l)[0][1:])
            ln = 24 + len(l)
            raw_packet = flat(
                PACKET_IN,
                ln,
                target_id,
                l.encode()
            )
            
            r.send(raw_packet)
        except:
            continue
        
    t.join()
    
         
if __name__ == "__main__":
    main()
    exit(0)