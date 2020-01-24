# Sample Script to check the windows event log via Icinga / Neamon / Nagios...

$log = "Microsoft-Windows-TerminalServices-RemoteConnectionManager"
$evtid = 1069
$minutes = 30
$counter = 1

if ((Get-EventLog -LogName $log -After (Get-Date).AddMinutes(-$minutes)  | Where-Object {$_.EventID -eq $evtid} | Select-Object -Property Source, EventID, InstanceId, Message ).count -gt $counter)
{
	echo "WARNING - EventID $evtid found within last $minutes minutes!" 
	$exitCode=1
}
	 
else {
    echo "OK - EventID $evtid not found within last $minutes minutes" 
	$exitCode=0
}
	 
exit ($exitCode)