This app was generated using `dotnet` CLI v6.0.100:
```
cd integration/testdata
dotnet new webapp -o /tmp/fde_dotnet_6

dotnet publish /tmp/fde_dotnet_6 --configuration Release --runtime ubuntu.18.04-x64 --self-contained false --output ./fde_6
```
