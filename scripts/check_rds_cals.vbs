' check_rds_cals.vbs
' Script to determine the usage of Microsoft 2016 RDS CALSs
' Author: Thorsten Eurich - ik4-sma
' Version 1.0
' -----------------------------------------------' 
Option Explicit
Dim objWMIService, objIssued, colIssued, strComputer, intUsed, colKeypack, objTotal, intTotal, intWarn, intCrit, strProductVersion

intWarn = 50
intCrit = 30
strProductVersion = "'Windows Server 2016'"

' On Error Resume Next
strComputer = "doctxrds"
intUsed = 0
intTotal = 0


' WMI connection zu Root CIM
Set objWMIService = GetObject("winmgmts:\\" _
& strComputer & "\root\cimv2")

' Total License
Set colKeypack = objWMIService.ExecQuery(_
"Select * from Win32_TSLicenseKeyPack where ProductVersion = " & strProductVersion)



' Loop through all Server 2016 Keypacks and sum them..
For Each objTotal in colKeypack
	' IssuedLicense 
	Set colIssued = objWMIService.ExecQuery(_
	"Select * from Win32_TSIssuedLicense where KeyPackId = " & objTotal.KeyPackId)

	' Loop through all issued Licenses and count them
	For Each objIssued in colIssued
	'	Wscript.Echo (objIssued.sIssuedToUser) &  " - " & objIssued.KeyPackId &  " - " & objIssued.LicenseId
		intUsed = intUsed + 1
	Next
	intTotal =  intTotal + objTotal.TotalLicenses
Next



' Prepare Output
' Critical, if less then 30 licenses left
If (intTotal - intUsed < intCrit) Then
    Wscript.Echo "CRITICAL : " & intUsed & " von " & intTotal & " RDS 2016 CALs in Benutzung|inUse=" & intUsed & ";" & intTotal - intWarn & ";" & intTotal - intCrit & ";0;" & intTotal
	WSCript.Quit(2)

' warning, if less then 50 licenses left
Elseif (intTotal - intUsed < intWarn) Then
	Wscript.Echo "WARNING : " & intUsed & " von " & intTotal & " RDS 2016 CALs in Benutzung|inUse=" & intUsed & ";" & intTotal - intWarn & ";" & intTotal - intCrit & ";0;" & intTotal
	WSCript.Quit(1)
' OK, output licenses
Else
	Wscript.Echo "OK: " & intUsed & " von " & intTotal & " RDS 2016 CALs in Benutzung|inUse=" & intUsed & ";" & intTotal - intWarn & ";" & intTotal - intCrit & ";0;" & intTotal
	WSCript.Quit(0)
End If