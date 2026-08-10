// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

using System;
using System.Collections.Generic;

namespace DbManager;

public partial class TAB_SOTTOTIPI
{
    public int IdSottotipo { get; set; }

    public string Descrizione { get; set; } = null!;

    public string TypeCode { get; set; } = null!;

    public string SubtypeCode { get; set; } = null!;

    public string Classname { get; set; } = null!;

    public string TypeValue { get; set; } = null!;

    public string? ExtraAttribute { get; set; }

    public string? ExtraValue { get; set; }

    public bool Ingresso { get; set; }

    public bool Uscita { get; set; }

    public bool Abilitato { get; set; }

    public string? Severity { get; set; }
}
