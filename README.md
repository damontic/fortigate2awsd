# fortigate2awsd
Send FortiOS 6.2 logs to AWS Cloduwatch Logs.
Runs a deamon which connects via SSH to a Fortigate Instance, gets its logs and sends them to Cloduwatch logs.

This was tested in:
- FortiGate-90D v6.0.5,build0268,190507
- FortiGate-VM64-AWSONDEMAND v6.2.1,build0932,190716 (GA)

The process executes `ssh` to connect to Fortigate and sends to Cloudwatch logs events.
To do that multiple streams must exist with the same post fix:
```
-{stream-prefix}_traffic
-{stream-prefix}_event
-{stream-prefix}_virus
-{stream-prefix}_webfilter
-{stream-prefix}_ips
-{stream-prefix}_emailfilter
-{stream-prefix}_anomaly
-{stream-prefix}_voip
-{stream-prefix}_dlp
-{stream-prefix}_app-ctrl
-{stream-prefix}_waf
-{stream-prefix}_dns
-{stream-prefix}_ssh
-{stream-prefix}_ssl
-{stream-prefix}_cifs
-{stream-prefix}_file-filter
```
Where `{stream-prefix}` is one of the program's parameters. This way Fortigate logs are categorized.
All those streams are stored in the same log group specified as another parameter.

# Usage
```bash
$ fortigate2awsd -h
Usage of fortigate2awsd:
  -dry-run
    	Set if you want to output messages to console. Useful for testing.
  -group string
    	Specify the log group where you want to send the logs
  -ip-port string
    	Specify the Fortigate ip and port to log to ip:port
  -password string
    	Specify the Fortigate ssh password
  -secret-manager string
    	Specify the AWS secrets manager secrets name to use as password
  -size int
    	Specify the number of events to send to AWS Cloudwatch. (default 100)
  -stream-prefix string
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