# fortigate2awsd
Send FortiOS 6.2 logs to AWS Cloduwatch Logs.
Runs a deamon which connects via SSH to a Fortigate Instance, gets its logs and sends them to Cloduwatch logs.

This was tested in:
- FortiGate-90D v6.0.5,build0268,190507
- FortiGate-VM64-AWSONDEMAND v6.2.1,build0932,190716 (GA)

The process executes `ssh` to connect to Fortigate and sends to Cloudwatch logs events.
To do that multiple log gruops must exist with the same post fix:
```
-{log-group-prefix}-traffic
-{log-group-prefix}-event
-{log-group-prefix}-virus
-{log-group-prefix}-webfilter
-{log-group-prefix}-ips
-{log-group-prefix}-emailfilter
-{log-group-prefix}-anomaly
-{log-group-prefix}-voip
-{log-group-prefix}-dlp
-{log-group-prefix}-app-ctrl
-{log-group-prefix}-waf
-{log-group-prefix}-dns
-{log-group-prefix}-ssh
-{log-group-prefix}-ssl
-{log-group-prefix}-cifs
-{log-group-prefix}-file-filter
```
Where `{log-group-prefix}` is one of the program's parameters. This way Fortigate logs are categorized.
Every log group will have a stream with the specified (as a process argument) name created.

# Usage
```bash
$ fortigate2awsd -h
Usage of fortigate2awsd:
  -dry-run
    	Set if you want to output messages to console. Useful for testing.
  -group-prefix string
    	Specify the log group prefix where you want to send the logs
  -ip-port string
    	Specify the Fortigate ip and port to log to ip:port
  -password string
    	Specify the Fortigate ssh password
  -period int
    	Specify the number of seconds to wait between logs category pushes. (default 300)
  -secret-manager string
    	Specify the AWS secrets manager secrets name to use as password
  -size int
    	Specify the number of events to send to AWS Cloudwatch. (default 100)
  -stream string
    	Specify the log stream where you want to send the logs
  -username string
    	Specify the Fortigate ssh username
  -verbose
    	Set if you want to be verbose.
  -version
    	Set if you want to see the version and exit.
```

# Run as a service
Two systemd unit files are provided.

- Make sure that the machine which is going to execute the daemon, has ssh connection to the Fortigate instance.
- Keep in mind that the binary `fortigate2awsd` must be available here `/usr/local/bin/`.
- Copy the `fortigate2awsd.service` unit file in the `/etc/systemd/system/` directory.
- Run `# systemctl daemon-reload`
- Run `# systemctl start fortigate2awsd.service`
- Run `# systemctl status fortigate2awsd.service