using DatexPub;
using DatexPub.Model;
using DbManager;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Options;

var builder = Host.CreateApplicationBuilder(args);
builder.Services.AddHostedService<Worker>();

builder.Services.AddOptions<ConfigData>().Bind(builder.Configuration.GetSection("ConfigData"));
builder.Services.AddSingleton(sp => sp.GetRequiredService<IOptions<ConfigData>>().Value);

var connectionString = builder.Configuration["ConfigData:DBConfig:DBSource"];
builder.Services.AddDbContext<postgresContext>(options => options.UseNpgsql(connectionString));

var host = builder.Build();
host.Run();
