# fortigate2awsd
Send FortiOS 6.2 logs to AWS Cloduwatch Logs.
Runs a deamon which connects via SSH to a Fortigate Instance, gets its logs and sends them to Cloduwatch logs.

This was tested in:
- FortiGate-90D v6.0.5,build0268,190507

The process executes `ssh` to connect to Fortigate and sends to Cloudwatch logs events every 10 new lines.

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
  -size int
    	Specify the number of events to send to AWS Cloudwatch. (default 10)
  -stream string
    	Specify the log stream where you want to send the logs
  -username string
    	Specify the Fortigate ssh username
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