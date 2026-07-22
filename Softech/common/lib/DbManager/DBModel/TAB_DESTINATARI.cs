using System;
using System.Collections.Generic;

namespace DbManager;

public partial class TAB_DESTINATARI
{
    public int IdDestinatario { get; set; }

    public string Descrizione { get; set; } = null!;

    public bool Invio { get; set; }

    public bool Ricezione { get; set; }

    public string SupplierInterno { get; set; } = null!;

    public string Supplier { get; set; } = null!;

    public string Traduttore { get; set; } = null!;

    public string Confidentiality { get; set; } = null!;

    public bool Abilitato { get; set; }

    public string? Url { get; set; }
}
