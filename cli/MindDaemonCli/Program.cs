using Grpc.Net.Client;
using MindDaemonCli.Proto;
using Spectre.Console;

const string DefaultServerUrl = "http://localhost:8080";

var serverUrl = Environment.GetEnvironmentVariable("MIND_DAEMON_SERVER_URL");
if (string.IsNullOrWhiteSpace(serverUrl))
{
    serverUrl = DefaultServerUrl;
}

AnsiConsole.Write(
    new FigletText("Mind Daemon")
        .LeftJustified()
        .Color(Color.Teal));

var exit = false;
while (!exit)
{
    var choice = AnsiConsole.Prompt(
        new SelectionPrompt<string>()
            .Title("Select an action")
            .AddChoices("Say hello", "Show server address", "Exit"));

    switch (choice)
    {
        case "Say hello":
            await SayHelloAsync(serverUrl);
            break;
        case "Show server address":
            AnsiConsole.Write(new Panel(serverUrl).Header("Server"));
            break;
        case "Exit":
            exit = true;
            break;
    }
}

static async Task SayHelloAsync(string serverUrl)
{
    var name = AnsiConsole.Ask("Name:", Environment.UserName);

    try
    {
        var reply = await AnsiConsole.Status()
            .Spinner(Spinner.Known.Dots)
            .StartAsync("Calling server...", async _ =>
            {
                using var channel = GrpcChannel.ForAddress(serverUrl);
                var client = new HelloService.HelloServiceClient(channel);

                return await client.SayHelloAsync(new HelloRequest { Name = name });
            });

        var table = new Table()
            .Border(TableBorder.Rounded)
            .AddColumn("Field")
            .AddColumn("Value")
            .AddRow("Server", serverUrl)
            .AddRow("Message", reply.Message);

        AnsiConsole.Write(table);
    }
    catch (Exception ex)
    {
        AnsiConsole.MarkupLine("[red]Unable to call the local server.[/]");
        AnsiConsole.MarkupLineInterpolated($"[grey]{ex.Message}[/]");
    }
}
