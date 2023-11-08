This app was built with .NET SDK v7.0.100. To recreate:
1. `dotnet new mvc --name fde_dotnet_7 --output /tmp/fde_dotnet_7`
2. `dotnet publish /tmp/fde_dotnet_7 --runtime ubuntu.18.04-x64 --self-contained false --output /tmp/fde_dotnet_7_built`
