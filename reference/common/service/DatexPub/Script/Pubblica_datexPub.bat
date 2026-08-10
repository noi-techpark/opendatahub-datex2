REM SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
REM
REM SPDX-License-Identifier: CC0-1.0

#
# SPDX-License-Identifier: CC0-1.0

docker stop datexpub
docker container prune --force

cd E:\Sviluppo\Mobilita\DatexBZ\common

docker build --no-cache -t datexpub -f service/DatexPub/Dockerfile .

docker tag datexpub 192.168.1.245:5000/datexpub
docker push 192.168.1.245:5000/datexpub

rem docker tag datexpub 192.168.2.174:5000/datexpub
rem docker push 192.168.2.174:5000/datexpub

rem docker tag datexpub 192.168.119.130:5000/datexpub
rem docker push 192.168.119.130:5000/datexpub

docker tag datexpub datexbz.softech-hub.com/datexpub
docker push datexbz.softech-hub.com/datexpub

pause
