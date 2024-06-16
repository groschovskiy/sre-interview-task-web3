{ pkgs, ... }: {
  channel = "stable-23.11";
  packages = [
    pkgs.go
    pkgs.air
  ];
  env = {};
  idx = {
    extensions = [
      "golang.go"
      "rangav.vscode-thunder-client"
    ];
    workspace = {
      onStart= {
        run-server = "air";
      };
    };
  };
}
