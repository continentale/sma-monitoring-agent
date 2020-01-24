# Performance Counter Name
$counterName = "\Prozessorinformationen(_Total)\Prozessorzeit (%)"

# Output messages
$ok = "OK: Processor time is: "
$warning = "WARNING: Processor time:"
$critical = "CRITICAL: Processor time:"


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

if ($value -gt 95)
{
  
    echo "$warning : $value"
    exit (1)
 
}
elseif ($value -gt 99)
{
        echo "$crtical : $value" 
        exit (2)
}
else 
{
    echo "$ok : $value" 
    exit (0)
}