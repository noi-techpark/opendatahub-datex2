// SPDX-FileCopyrightText: 2026 2026 NOI Techpark <digital@noi.bz.it>
// SPDX-FileCopyrightText: 2026, 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

using System;
using System.Collections.Generic;
using Microsoft.EntityFrameworkCore;

namespace DbManager;

public partial class postgresContext : DbContext
{
    public postgresContext(DbContextOptions<postgresContext> options)
        : base(options)
    {
    }

    public virtual DbSet<TAB_DESTINATARI> TAB_DESTINATARI { get; set; }

    public virtual DbSet<TAB_PARAMETRI> TAB_PARAMETRI { get; set; }

    public virtual DbSet<TAB_SOTTOTIPI> TAB_SOTTOTIPI { get; set; }

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.Entity<TAB_DESTINATARI>(entity =>
        {
            entity.HasKey(e => e.IdDestinatario).HasName("TAB_DESTINATARI_pkey");

            entity.Property(e => e.IdDestinatario).ValueGeneratedNever();
            entity.Property(e => e.Confidentiality).HasMaxLength(200);
            entity.Property(e => e.Descrizione).HasMaxLength(200);
            entity.Property(e => e.Supplier).HasMaxLength(20);
            entity.Property(e => e.SupplierInterno).HasMaxLength(20);
            entity.Property(e => e.Traduttore).HasMaxLength(200);
            entity.Property(e => e.Url).HasMaxLength(1000);
        });

        modelBuilder.Entity<TAB_PARAMETRI>(entity =>
        {
            entity.HasKey(e => e.IdParametro).HasName("TAB_PARAMETRI_pkey");

            entity.Property(e => e.IdParametro).ValueGeneratedNever();
            entity.Property(e => e.Modulo).HasMaxLength(100);
            entity.Property(e => e.Note).HasMaxLength(1000);
            entity.Property(e => e.Parametro).HasMaxLength(300);
            entity.Property(e => e.Valore).HasMaxLength(4000);
        });

        modelBuilder.Entity<TAB_SOTTOTIPI>(entity =>
        {
            entity.HasKey(e => e.IdSottotipo).HasName("TAB_SOTTOTIPI_pkey");

            entity.Property(e => e.IdSottotipo).ValueGeneratedNever();
            entity.Property(e => e.Classname).HasMaxLength(200);
            entity.Property(e => e.Descrizione).HasMaxLength(200);
            entity.Property(e => e.ExtraAttribute).HasMaxLength(200);
            entity.Property(e => e.ExtraValue).HasMaxLength(200);
            entity.Property(e => e.Severity).HasMaxLength(50);
            entity.Property(e => e.SubtypeCode).HasMaxLength(200);
            entity.Property(e => e.TypeCode).HasMaxLength(200);
            entity.Property(e => e.TypeValue).HasMaxLength(200);
        });

        OnModelCreatingPartial(modelBuilder);
    }

    partial void OnModelCreatingPartial(ModelBuilder modelBuilder);
}
