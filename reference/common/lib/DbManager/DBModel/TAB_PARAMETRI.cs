// SPDX-FileCopyrightText: 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

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
