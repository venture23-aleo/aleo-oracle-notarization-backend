**Aleo Oracle program id**: [veru\_oracle\_v2.aleo](https://aleoscan.io/program?id=veru_oracle_v2.aleo) 

**To retrieve the token price from aleo program:** 

We need to call the `sgx_attested_data` method of the oracle program with respective hash for each token:

- ALEO : *69545840751771345196310203380793058459u128*  
- USDC : *142709967224662814311225757187928108124u128*  
- USDT : *220938201505672543703232741194538576644u128*  
- BTC : *148085669420309002516829472376152452678u128*  
- ETH : *153473196284127288038645336198866717144u128*

The attestation data uses a precision of `6(configurable)`,  so the returned value should be divided by `10^6`.

**Note: Since we are using the ALEO, USDC, and USDT token prices in Verulend, their prices are updated every 5 minutes. The update frequency for BTC and ETH token prices is currently set to 2 hours.**

For example, to fetch the price of **ALEO**, use its request hash.

![](./images/aleo-token-price-screenshot.png)


Example:  
 `127789 / 1000000 = 0.127789`
