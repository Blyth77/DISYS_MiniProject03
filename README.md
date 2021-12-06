# DISYS_MiniProject03 Solution#2
=======================================================================================================
NOTE: 
The clean_log.go can be run to clean the log folders. To run it, give it an argument N. N being the number of replicas you wish the client's to "know", this is hardcoded to a list all clients know. THIS IS NOT PART OF THE SYSTEM, only a helper program (We know it is violating the definition of distributed systems! The program can be used without it!)

If clean_logs is not run, 4 replica ports will be available by default:
    - port:  6001,
             6002,
             6003,
             6004

To add more run the clean_logs.go with N as argument. N being the number of available ports. NOTE: not all ports needs to be used for the program to work. Only 2 are required to test the program!

-------------------------------------------------------------------------------------------------------
Start-up instructions:
-------------------------------------------------------------------------------------------------------
START UP REPLICA: start by running N number of replicas from different terminals:
    - go run replicamanager/replicaManager.go <PORT> 
    - go run replicamanager/replicaManager.go <6001> 
    - go run replicamanager/replicaManager.go <6002> 
    .
    .
    - go run replicamanager/replicaManager.go <6000 + N> 
    (only the number "6001" as an argument, no ":", and only one replica on each port)

NEXT: Follow the instructions, and initiate an auction, by choosing an item and a timespan (seconds):

    Please enter an item and a timespan(seconds) for the auction:    
        "dog 60"     
        (item: dog, time: 60 seconds)

This configuration must be done on EVERY REPLICA! The auction won't start before a client puts in the first bid, so there is plenty of time to do this. If the messages are not identical, unexpected results will follow! 

When the replica is running, there are two options:
    - Type "kill" to simulate a failure. 
    - Type "new" to start a new auction (cannot be done before the previous auction is finished).

-------------------------------------------------------------------------------------------------------
START UP CLIENT: When the replicas are in auction-mode (have an item and a timespan), we can start up N number of clients:
    - go run client/client.go <PORT>
    - go run client/client.go 3001
    ...

    (Use any ports (replica ports excluded), no repetitions)

OPTIONS:
    Type "bid <amount>"   ex. "bid 100" to bid 100.
    Type "query"  to se current highest bidder. It will also show if the auction has finished.
    Type "quit" to exit. 
    Type "help" for instructions. 

After an auction you can start another auction. But you have to enter the configuration into every replica. The client will discover the new auction by bidding something. It will say if a new auction is available. 

=======================================================================================================



