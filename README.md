# beatles-client-lib  
a vpn client lib of beatles   

//todo list  
[ok]1. bootstrap msg from github    
[ok]2. support paclist  
[ok]3. client -> miner stream  
4. client setting [mac,win,linux  ]  [mac ok]   
[ok] 5. purchase from eth         
[ok] 6. use purchase result to get license    
[ok] 7. flush miners                        
[ok] 8. start vpn   
9. miners health param [throughout, time delay, session per second]  
10. support import license  
11. support export license  
12. support export wallet  as a cipher text or  eth privacy or trx privacy  
13. support import from matamask or tronmask  

============================  
mac [first use] :  
1. ./btlclient daemon  
2. ./btlclient start  
3. ./btlclient eth  
	Eth Address: 0x778196e979839Fb5849BD2A91038f5a2C04e7e82  
	Beatles Address: tg2KWECaThqy1uZstFyGoKBWXbkUmHaEoRyZ6Web1fHC1gfu  
	Eth Balance: 0  
4. transfer some ropsten eth to [Eth Address]  
5. ./btlclient eth price -m 12 
6. ./btlclient eth buy -e "test@abc.com" -n "test" -c "11223344" 
7. ./btlclient eth license  
8. ./btlclient miner flush  
9. ./btlclient miner  
   [  
 	{  
 		"ipv_4_addr": "45.32.52.199",  
 		"port": 47911,  
 		"location": "jp-tokyo",  
 		"miner_id": "tg2KdqCZGCEFQxYVikUa6syBjJE6qXm7BzfUyGEzM28VJtoN"  
 	},  
 	{  
 		"ipv_4_addr": "34.96.156.219",  
 		"port": 46637,  
 		"location": "hk-lanbery",  
 		"miner_id": "tg2KYebW3jpZbqthZUSiKwhVnohKMQfSTMtn3PZqk5ZH5avS"  
 	}  
  ]  
10. ./btlclient start vpn -m0  


mac[if u have license]  
1. ./btlclient daemon  
2. ./btlclient start  
3. ./btlclient start vpn -m1  


change mode  
   ./btlclient vpnmode -m [1,0]  
choose other miner  
   ./btlclient stop vpn  
   ./btlclient start vpn -m1  

