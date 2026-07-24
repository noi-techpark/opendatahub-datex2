// SPDX-FileCopyrightText: 2026 2026 NOI Techpark <digital@noi.bz.it>
// SPDX-FileCopyrightText: 2026, 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

# SPDX-FileCopyrightText: 2026 2026 NOI Techpark <digital@noi.bz.it>
#
# SPDX-License-Identifier: CC0-1.0

using System;
using System.Collections.Generic;
using System.Text.Json.Serialization;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Supporto
{
	public class RootResponse
	{
		public int TotalResults { get; set; } = 0;
		public int TotalPages { get; set; } = 0;
		public int CurrentPage { get; set; } = 0;
		public string PreviousPage { get; set; } = string.Empty;
		public string NextPage { get; set; } = string.Empty;
		public string Seed { get; set; } = string.Empty;
		public List<Item> Items { get; set; } = new();
	}

	public class Item
	{
		public string Id { get; set; } = string.Empty;
		public bool Active { get; set; } = false;
		public DateTime StartTime { get; set; } = DateTime.MinValue;
		public DateTime? EndTime { get; set; }
		public string Shortname { get; set; } = string.Empty;
		public string Source { get; set; } = string.Empty;
		public DateTime FirstImport { get; set; } = DateTime.MinValue;
		public DateTime LastChange { get; set; } = DateTime.MinValue;

		// Dizionari inizializzati per evitare null ref
		public Dictionary<string, DetailInfo> Detail { get; set; } = new();
		public Dictionary<string, GeoLocation> Geo { get; set; } = new();
		public Dictionary<string, Dictionary<string, string>> Mapping { get; set; } = new();

		public List<RelatedContent> RelatedContent { get; set; } = new();

		public Meta _Meta { get; set; } = new();

		public LicenseInfo LicenseInfo { get; set; } = new();
		public List<string> HasLanguage { get; set; } = new();
		public List<string> TagIds { get; set; } = new();
		public AdditionalProperties AdditionalProperties { get; set; } = new();
	}

	public class DetailInfo
	{
		public string BaseText { get; set; } = string.Empty;
		public string Title { get; set; } = string.Empty;
		public string Language { get; set; } = string.Empty;
	}

	public class GeoLocation
	{
		public string Gpstype { get; set; } = string.Empty;
		public double? Latitude { get; set; }
		public double? Longitude { get; set; }
		public double? Altitude { get; set; }
		public string AltitudeUnitofMeasure { get; set; } = string.Empty;
		public string Geometry { get; set; } = string.Empty;
		public bool Default { get; set; } = false;
	}

	public class Meta
	{
		public string Id { get; set; } = string.Empty;
		public string Type { get; set; } = string.Empty;
		public DateTime LastUpdate { get; set; } = DateTime.MinValue;
		public string Source { get; set; } = string.Empty;
		public bool Reduced { get; set; } = false;
		public UpdateInfo UpdateInfo { get; set; } = new();
	}

	public class UpdateInfo
	{
		public string UpdatedBy { get; set; } = string.Empty;
		public string UpdateSource { get; set; } = string.Empty;
		public int Revision { get; set; } = 0;
		public List<UpdateHistory> UpdateHistory { get; set; } = new();
	}

	public class AdditionalProperties
	{
		public EchargingProperties EchargingDataProperties { get; set; } = new();
		public ActivityLtsProperties ActivityLtsDataProperties { get; set; } = new();
		public GastronomyProperties GastronomyLtsDataProperties { get; set; } = new();
		// Dictionary per proprietà dinamiche non strutturate
		public Dictionary<string, object> RoadIncidentProperties { get; set; } = new();
	}

	public class EchargingProperties
	{
		public int Capacity { get; set; } = 0;
		public string State { get; set; } = "UNAVAILABLE";
		public List<string> ChargingPistolTypes { get; set; } = new();
		public ParkingSpace CarParkingSpaceNextToEachOther { get; set; } = new();
		public bool Covered { get; set; } = false;
	}

	public class ParkingSpace
	{
		public bool Flat { get; set; } = true;
		public double Gradient { get; set; } = 0;
		public string Pavement { get; set; } = string.Empty;
		public double Width { get; set; } = 0;
		public double Length { get; set; } = 0;
	}

	public class ActivityLtsProperties
	{
		public double AltitudeDifference { get; set; } = 0.0;
		public bool IsOpen { get; set; } = false;
		public Ratings Ratings { get; set; } = new();
		public List<string> Exposition { get; set; } = new();
	}

	public class Ratings
	{
		public string Stamina { get; set; } = string.Empty;
		public string Difficulty { get; set; } = string.Empty;
	}

	public class GastronomyProperties
	{
		public int MaxSeatingCapacity { get; set; } = 0;
		public List<CategoryCode> CategoryCodes { get; set; } = new();
		public List<DishRate> DishRates { get; set; } = new();
	}

	public class DishRate
	{
		public string Id { get; set; } = string.Empty;
		public double MinAmount { get; set; } = 0;
		public double MaxAmount { get; set; } = 0;
		public string CurrencyCode { get; set; } = string.Empty;
	}

	public class LicenseInfo
	{
		public string License { get; set; } = "CCO";
		public string LicenseHolder { get; set; } = string.Empty;
		public bool ClosedData { get; set; } = false;
	}

	public class RelatedContent
	{
		public string Id { get; set; } = string.Empty;
		public string Type { get; set; } = string.Empty;
		public string Self { get; set; } = string.Empty;
	}

	public class UpdateHistory
	{
		public DateTime LastUpdate { get; set; } = DateTime.MinValue;
		public string UpdateSource { get; set; } = string.Empty;
		public string UpdatedBy { get; set; } = string.Empty;
	}

	public class CategoryCode
	{
		public string Id { get; set; } = string.Empty;
		public string Shortname { get; set; } = string.Empty;
	}
}
