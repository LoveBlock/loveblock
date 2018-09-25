# LoveBlock

LoveBlock is the decentralized database for Dating - DDD and LoveBlock.one (LB) is the most innovative blockchain technology solution for online dating around the world! 
The LB blockchain infrastructure and dating platform will empore Blockchain and Smart Contracts to deal with the major problems in online dating: 
LoveBlock solves these key issues related to online dating, such as fraudulent activity, motivation to acquire tokens to use services, security using block chains. 
Based out of the laboratory in Singapore, the LB team works closely together with Luxy (www.onluxy.com), a highly successful dating application with more than 2 million users worldwide. 
In LB, our company's unique verification system makes it easy to identify fraud. Information is stored on the LoveBlock Blockchain, encrypted, and can share data with other DApps. 
By doing this, fraudsters will have no place to hide and can be eliminated permanently. 


## Installation Instructions
Setup build tools
- Install `golang`, ensure the Go version is 1.7(or any later version).
- Install `git`
- Install `mingw` if you use windows to build, cause GCC is required
- Set up the Path environment variable `GOPATH`
- Make source code directory in `GOPATH` and download source code into it
    ```
    mkdir src\github.com\LoveBlock
    git clone https://github.com/LoveBlock/loveblock src\github.com\LoveBlock\loveblock
    ```
- Compile loveblock
    ```
    cd src\github.com\LoveBlock\loveblock
    go install -v ./loveblock
    ```



## Running loveblock
Start up LoveBlock's built-in interactive JavaScript console, (via the trailing `console` subcommand) through which you can invoke all official `network` methods. You can simply interact with the LoveBlock network; create accounts; transfer funds; deploy and interact with contracts. To do so:
```
$ loveblock console
```
This command will start a node to sync block datas from LoveBlock network. The `console` subcommand is optional and if you leave it out you can always attach to an already running node with `loveblock attach`.

Run with specific data directory
```
$ loveblock console --datadir=path/to/custom/data/folder
```



## Operating a private network
### Defining the private genesis state
First, you'll need to create the genesis state of your networks, which all nodes need to be aware of and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):
```json
{
  "config": {
    "chainId": 0
  },
  "alloc": {
    "0x0000000000000000000000000000000000000001": {
      "balance": "11111111111"
    },
    "0x0000000000000000000000000000000000000002": {
      "balance": "222222222"
    }
  },
  "nonce": "0x0000000000000613"
}
```
- `chainId` You can create a new block chain network totally separated with LoveBlock
- `alloc` Pre-fund accounts
- `nonce` Change the nonce to some random value so you prevent unknown remote nodes from being able to connect to you

With the genesis state defined in the above JSON file, you'll need to initialize every loveblock node with it prior to starting it up to ensure all blockchain parameters are correctly set:
```
loveblock init path/to/genesis.json
```

### Running star nodes
Star nodes confirm transactions and produce blocks.
1. Run loveblock with `console` command and record the `enode` address in log. (e.g. `enode://79ca6a10e28134087cce01c6b5f4dd54b7bb689c93b39aa29fb890f49c408589983c7a354f2c5d5c70072b7670153d3ac121e566c7ab6298dc53e201696da084@[::]:30303`)
2. Create account, enter the password and record address.
```
personal.newAccount()
```
3. Shutdown node. Fill account and enode address into `starlist` file in data directory. For Example:
```
0xc858fe1d6b73c0ec640b3b2469c40278c8adbb5a 79ca6a10e28134087cce01c6b5f4dd54b7bb689c93b39aa29fb890f49c408589983c7a354f2c5d5c70072b7670153d3ac121e566c7ab6298dc53e201696da084
```
  Every witness a line.
4. Run loveblock with star mode and start to mine
```
loveblock --nodemode star --mine <other flags>
```

### Starting up your satellite nodes
Satellite node is the default mode
1. Start every subsequent loveblock node with `console` command. It will probably also be desirable to keep the data directory of your private network separated, so do also specify a custom --datadir flag.
```
loveblock console --datadir=path/to/custom/data/folder
```
2. connect star nodes ( or other satellite nodes ) with `enode_address`
```
admin.addPeer("enode://79ca6a10e28134087cce01c6b5f4dd54b7bb689c93b39aa29fb890f49c408589983c7a354f2c5d5c70072b7670153d3ac121e566c7ab6298dc53e201696da084@[::]:30306")
```


## Thank you Ethereum
Ethereum is an outstanding open source blockchain project, with Turing complete virtual machine, convenient account data storage mechanism, and is a proven mature and stable system. However, Ethereum also has obvious problems of low throughput and low consensus efficiency. In order to create prototype rapidly and fast verify our innovative consensus algorithms and application scenarios, LoveBlock will perform in-depth optimization based on Ethereum 1.8.3, gradually replacing its consensus mechanism, introducing new smart contract language, and transforming the account system. And we will rewrite all code at July, 2018.

Thanks to all contributors of the Ethereum Project.
