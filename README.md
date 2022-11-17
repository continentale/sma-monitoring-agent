# sma-monitoring-agent
DEPRECATED: this agent is deprecated and will no longer be maintained. please follow the new agent on https://github.com/continentale/monitoring-agent

SMA-MonitoringAgent is a Windows based agent written in Golang to monitor a windows system. The main feature is collection 
basic system information such as CPU, disk and memory usage, running processes and services and to present the collected data via REST-API.

It can be easily extended via scripts straightforward.

## Getting started
To get started you have to download the binary. You can download it from the github release page. 
https://github.com/continentale/sma-monitoring-agent/releases/tag/v1.1.3

Then unzip it and you are ready to go.

## First calls to the API
For a first look its enough to view the results on a browser before you install the corresponding check. 
To view the results in the browser you navigate to the server running the program (or localhost) with the specified port (defaults: 10240) and paste the url of the metrics you want. 

The API-URLS:
* /api/diskusage
* /api/cpuusage
* /api/cpuusagebycore
* /api/memoryusage
* /api/processlist
* /api/services
* /api/systeminfo
* /api/exec
* /api/version

So if you want to see the version you navigate to e.G. http://localhost:10240/api/version and the output should can be 
```bash
{"Version":"1.0.0","BuildTime":"2019-08-01 UTC","GitHash":"e5b641847481bcdd7a19a01d355ab700e22ee7f3"}
```

Feel free to navigate to the other urls.

## Configuration
The configuration is placed in the agent.ini config file. The program searches the file in a environment variable called AGENT_INI_PATH.

Possible configuration:

| section  |     key     | datatype | description                                                                                         |
| -------- | :---------: | -------- | --------------------------------------------------------------------------------------------------- |
| server   |  protocol   | string   | protocol on which the server runs. [http or https ]                                                 |
| server   | certificate | string   | Path to the certificate for a https-Endpoint                                                        |
| server   | privatekey  | string   | Path to the Key for a https-Endpoint                                                                |
| server   |    port     | int      | Port on which the server runs                                                                       |
| server   |   secret    | string   | Specify a secret which is required before accessing the API                                         |
| commands |      *      | string   | Give your command a name and a path to call custom scripts from the api and get back a return value |


## Wiki
If you want to get started and get a more detailed impression of this project, then look at our wiki. [Getting started](https://github.com/continentale/sma-monitoring-agent/wiki/Getting-started)