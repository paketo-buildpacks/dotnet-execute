This app was generated using `dotnet` CLI v6.0.100:
```
cd integration/testdata
dotnet new blazorserver -o /tmp/blazor_6

dotnet publish /tmp/blazor_6 --configuration Release --runtime ubuntu.18.04-x64 --self-contained true --output ./self_contained_executable_6
```
