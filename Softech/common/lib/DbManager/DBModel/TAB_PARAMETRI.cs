using System;
using System.Collections.Generic;

namespace DbManager;

public partial class TAB_PARAMETRI
{
    public int IdParametro { get; set; }

    public string Modulo { get; set; } = null!;

    public string Parametro { get; set; } = null!;

    public string? Valore { get; set; }

    public string? Note { get; set; }
}
