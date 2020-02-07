@echo off
echo Testet Commands %1 %2 %3 %4 {{"counter":7, "temporary": "hello World"}}
@ping localhost -n 3 > NUL
exit /B 0