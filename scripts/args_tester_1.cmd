@echo off
echo OK: just a test message %1 %2 %3 %4
@ping localhost -n 3 > NUL
exit /B 0

