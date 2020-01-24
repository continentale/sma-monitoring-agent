## Sample Performance Counter script
## This script is intended to demonstrate the usage of collecting performance counter values via SMA-Monitoring-Agent
## we collect a procentual value and check, if it exceeds our given warning / critcal status
## @Author: Thorsten Eurich, ik4-sma


# Performance Counter Name
$counterName = "\Prozessorinformationen(_Total)\Prozessorzeit (%)"

# Output messages
$ok = "OK: Processor time is: "
$warning = "WARNING: Processor time:"
$critical = "CRITICAL: Processor time:"
$cn = "processor"
$warn = "95"
$crit = "99"


# Get PerfCounter value 
$total = @()

Get-Counter -Counter $counterName -SampleInterval 1 -MaxSamples 1 |
    Select-Object -ExpandProperty countersamples | % {
        $object = New-Object psobject -Property @{
            CookedValue = $_.CookedValue
        }

            $total += $object
    }

## Calculate average, we just have a single value so this is basically used to format the output
$value = ($total| Measure-Object -Average CookedValue).Average
$value = [math]::Round($value, 2)

# if $value is greater than $warn
if ($value -gt $warn)
{
  
    echo "$warning : $value|$cn=$value;$warn;$crit;;"
    exit (1)
 
}
# if $value is greater than $crit
elseif ($value -gt $crit)
{
        echo "$crtical : $value|$cn=$value;$warn;$crit;;"
        exit (2)
}
# not a warning, not a critical? then exit with ok
else 
{
    echo "$ok : $value|$cn=$value;$warn;$crit;;"
    exit (0)
}