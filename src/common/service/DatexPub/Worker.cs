// SPDX-FileCopyrightText: 2026 2026 NOI Techpark <digital@noi.bz.it>
// SPDX-FileCopyrightText: 2026, 2026 NOI Techpark <digital@noi.bz.it>
//
// SPDX-License-Identifier: CC0-1.0

using DbManager;
using DbManager.Utility;
using DatexPub.Model;
using RestSharp;
using Supporto;
using System.Reflection;
using Newtonsoft.Json;
using System.Collections.Generic;
using DatexModelOrig;
using NLog;
using NLog.Web;

namespace DatexPub
{
	public class Worker : BackgroundService
	{
		#region DATA
		static Logger log = NLogBuilder.ConfigureNLog("NLog.config").GetLogger("Worker");
		private readonly ILogger<Worker> logger;
		private ConfigData configData;
		private List<string> openDataHubTags = new List<string>
		{
			"traffic-event:accident",
			"traffic-event:animal-on-road",
			"traffic-event:caution",
			"traffic-event:closure",
			"traffic-event:congestion",
			"traffic-event:event",
			"traffic-event:hindrance",
			"traffic-event:prohibition",
			"traffic-event:road-condition",
			"traffic-event:road-work",
			"traffic-event:speed-camera",
			"traffic-event:weather-related"
		};
		#endregion

		public Worker(ILogger<Worker> log, ConfigData configData)
		{
			logger = log;
			this.configData = configData;
			DbConfiguration.SetConfiguration(this.configData.DBConfig.DBSource);
		}

		protected override async Task ExecuteAsync(CancellationToken stoppingToken)
		{
			log.Info("----------------- START -----------------");

			//log.Info("DB Connect: " + this.configData.DBConfig.DBSource);

			while (!stoppingToken.IsCancellationRequested)
			{
				log.Debug("Elaborazione");

				// recupero eventi da open data hub
				Task<string> task = GetEventiOpenDataHub();
				task.Wait();
				var eventi = task.Result;
				if (!string.IsNullOrEmpty(eventi))
				{
					RootResponse? eventiStruct = JsonConvert.DeserializeObject<RootResponse>(eventi);
					if (eventiStruct != null)
					{
						log.Info("Eventi ritornati: " + eventiStruct.TotalResults);

						List<Datex2Event>? eventiAttivi = PreparaEventi(eventiStruct);
						if (eventiAttivi != null)
						{
							TraduciEventi(eventiAttivi);
						}
					}
					else
					{
						log.Warn("Risposta non deserializzata");
					}
				}

				await Task.Delay(configData.TimeoutElaborazione, stoppingToken);
			}
		}

		private async Task<string> GetEventiOpenDataHub()
		{
			try
			{
				log.Trace("in");

				var options = new RestClientOptions();
				var client = new RestClient(options);
				var request = new RestRequest(configData.UrlEventiBolzano);
				var response = await client.GetAsync(request);

				if (response != null)
				{
					if (response.IsSuccessful && response.IsSuccessStatusCode && response.Content != null)
					{
						log.Debug("Risposta da " + configData.UrlEventiBolzano);
						return response.Content;
					}
				}

				log.Warn("Risposta NULL da " + configData.UrlEventiBolzano);
				return "";
			}
			catch (Exception ex)
			{
				LogException(MethodBase.GetCurrentMethod(), ex);
				return "";
			}
			finally
			{
				log.Trace("out");
			}
		}

		private List<Datex2Event>? PreparaEventi(RootResponse eventiStruct)
		{
			try
			{
				log.Trace("in");

				List<Item>? bolzano = FiltraEventiBolzano(eventiStruct.Items);
				if (bolzano == null)
				{
					log.Warn("Errore durante filtro eventi di Bolzano");
					return null;
				}

				List<Item>? attuali = FiltraEventiAttuali(bolzano);
				if (attuali == null)
				{
					log.Warn("Errore durante filtro eventi attuali");
					return null;
				}

				// conversione eventi attivi nella struttura interna
				List<Datex2Event> eventi = MapItemsToDatex2(attuali);

				return eventi;
			}
			catch (Exception ex)
			{
				LogException(MethodBase.GetCurrentMethod(), ex);
				return new List<Datex2Event>();
			}
			finally
			{
				log.Trace("out");
			}
		}

		public List<Item>? FiltraEventiBolzano(List<Item> eventi)
		{
			try
			{
				log.Trace("in");

				List<Item> bolzano = eventi.Where(x => x._Meta.Source == "PROVINCE_BZ").ToList();
				return bolzano;
			}
			catch (Exception ex)
			{
				LogException(MethodBase.GetCurrentMethod(), ex);
				return null;
			}
			finally
			{
				log.Trace("out");
			}
		}

		public List<Item>? FiltraEventiAttuali(List<Item> eventi)
		{
			try
			{
				log.Trace("in");

				var oraCorrente = DateTime.Now;
				var oggi = oraCorrente.Date;

				// Categorie definite come "turbative sul breve periodo"
				var categorieBrevePeriodo = new HashSet<string>
				{
					"traffic-event:speed-camera",
					"traffic-event:congestion",
					"traffic-event:accident",
					"traffic-event:event",
					"traffic-event:animal-on-road",
					"traffic-event:weather-related"
				};

				List<Item> attuali = eventi.Where(ev =>
				{
					// Verifica se l'evento ha almeno una categoria a breve periodo
					bool brevePeriodo = ev.TagIds.Any(tag => categorieBrevePeriodo.Contains(tag));

					if (brevePeriodo)
					{
						// se ha una data di fine, non deve essere nel passato
						if (ev.EndTime.HasValue)
							return ev.EndTime.Value > oraCorrente;

						// se non ha data di fine, deve essere iniziato oggi
						return ev.StartTime.Date == oggi;
					}
					else
					{
						// se non ha data di fine, è sempre valido
						if (!ev.EndTime.HasValue)
							return true;

						// se ha data di fine, l'ora corrente deve essere compresa nel periodo
						return oraCorrente >= ev.StartTime && oraCorrente <= ev.EndTime.Value;
					}
				}).ToList();

				return attuali;
			}
			catch (Exception ex)
			{
				LogException(MethodBase.GetCurrentMethod(), ex);
				return null;
			}
			finally
			{
				log.Trace("out");
			}
		}

		public List<Datex2Event> MapItemsToDatex2(List<Item> eventi)
		{
			try
			{
				List<Datex2Event> ret = new List<Datex2Event>();

				if (eventi == null)
					return ret;

				using (postgresContext? db = DBHelper.Connect())
				{
					if (db != null)
					{
						List<TAB_SOTTOTIPI> sottotipi = DbUtility.Sottotipi_LoadAll(db);

						foreach (Item item in eventi)
						{
							Datex2Event eve = new Datex2Event();

							// identificativi
							eve.Id = item.Id.Replace(configData.PrefissoIdentificativi, "");
							eve.Situazione = eve.Id;

							Dictionary<string, string> providerProvinceBz = item.Mapping["ProviderProvinceBz"];
							if (providerProvinceBz != null)
							{
								eve.Reference = providerProvinceBz["Id"];
							}

							eve.Version = 1;
							if (item._Meta != null && item._Meta.UpdateInfo != null && item._Meta.UpdateInfo.UpdateHistory != null)
							{
								if (item._Meta.UpdateInfo.UpdateHistory.Count > 0)
									eve.Version = item._Meta.UpdateInfo.UpdateHistory.Count;
							}

							// sorgente dati
							eve.SourceCountry = "it";
							eve.SourceType = "roadAuthorities";
							eve.SourceReliable = true;
							if (item._Meta != null)
							{
								eve.SourceName = item._Meta.Source;
							}

							// dati temporali
							eve.StartTime = item.StartTime;
							eve.EndTime = item.EndTime;
							eve.CreationTime = item.FirstImport;
							eve.VersionTime = item.LastChange;

							// categoria evento
							string? tagIndividuato = item.TagIds.FirstOrDefault(tag => openDataHubTags.Contains(tag));
							if (tagIndividuato != null)
							{
								eve.Categoria = tagIndividuato;
							}
							else
							{
								log.Warn("Categoria non individuata per: " + item.Id);
								continue;
							}

							// dati spaziali
							if (item.Geo != null)
							{
								eve.Latitude = item.Geo["position"].Latitude ?? 0;
								eve.Longitude = item.Geo["position"].Longitude ?? 0;
							}

							// descrizioni
							eve.GeneralPublicCommentType = "description";

							if (item.Detail.ContainsKey("it") && item.Detail["it"] != null)
							{
								eve.GeneralPublicCommentIt = item.Detail["it"]?.Title ?? string.Empty;
								if (!string.IsNullOrEmpty(item.Detail["it"].BaseText))
									eve.GeneralPublicCommentIt += " " + item.Detail["it"].BaseText;
							}

							if (item.Detail.ContainsKey("en") && item.Detail["en"] != null)
							{
								eve.GeneralPublicCommentEn = item.Detail["en"]?.Title ?? string.Empty;
								if (!string.IsNullOrEmpty(item.Detail["en"].BaseText))
									eve.GeneralPublicCommentEn += " " + item.Detail["en"].BaseText;
							}

							if (item.Detail.ContainsKey("de") && item.Detail["de"] != null)
							{
								eve.GeneralPublicCommentDe = item.Detail["de"]?.Title ?? string.Empty;
								if (!string.IsNullOrEmpty(item.Detail["de"].BaseText))
									eve.GeneralPublicCommentDe += " " + item.Detail["de"].BaseText;
							}

							// severity (basata su categoria)
							TAB_SOTTOTIPI? sot = sottotipi.FirstOrDefault(x => x.SubtypeCode == eve.Categoria);
							if (sot != null && sot.Severity != null)
								eve.OverallSeverity = sot.Severity;

							// valori di default
							eve.AreaOfInterest = "regional";
							eve.Confidentiality = "noRestriction";
							eve.InformationStatus = "real";
							eve.Urgency = "normalUrgency";
							eve.Probability = "certain";

							ret.Add(eve);
						}
					}
				}

				return ret;
			}
			catch (Exception ex)
			{
				LogException(MethodBase.GetCurrentMethod(), ex);
				return new List<Datex2Event>();
			}
			finally
			{
				log.Trace("out");
			}
		}

		private void TraduciEventi(List<Datex2Event> eventi)
		{
			try
			{
				log.Trace("in");

				using (postgresContext? db = DBHelper.Connect())
				{
					if (db != null)
					{
						TraduttoreSituation trad = new TraduttoreSituation(db);

						// supplier interno
						trad.PublicationCreator = "";
						TAB_PARAMETRI? parSupplier = DbUtility.Parametri_Load("NodoDatex", "SupplierInterno", db);
						if (parSupplier != null && parSupplier.Valore != null)
							trad.PublicationCreator = parSupplier.Valore;

						// destinatari
						List<TAB_DESTINATARI> destinatari = DbUtility.Destinatari_LoadAll(db);

						// elaboro i destinatari
						foreach (TAB_DESTINATARI dest in destinatari)
						{
							// genero xml datex II
							string xml = trad.GetD2LogicalModelXml(eventi, dest.IdDestinatario);

							// salvo pubblicazione
							string directory = configData.PathPubblicazioni + "/Invio/" + dest.Supplier;
							SaveToFile(xml, directory, "SituationPublication.xml");

							// log
							string msg = "Pubblicato per destinatario: " + dest.Descrizione + " (" + eventi.Count + " eventi)";
							log.Info(msg);
							logger.LogInformation(msg);
						}
					}
				}
			}
			catch (Exception ex)
			{
				LogException(MethodBase.GetCurrentMethod(), ex);
			}
			finally
			{
				log.Trace("out");
			}
		}

		private void SaveToFile(string xmlString, string directory, string fileName)
		{
			StreamWriter? streamWriter = null;
			try
			{
				log.Trace("in");

				Directory.CreateDirectory(directory);

				FileInfo xmlFile = new FileInfo(directory + "/" + fileName);
				streamWriter = xmlFile.CreateText();
				streamWriter.WriteLine(xmlString);
				streamWriter.Close();
			}
			catch (Exception ex)
			{
				LogException(MethodBase.GetCurrentMethod(), ex);
			}
			finally
			{
				if ((streamWriter != null))
				{
					streamWriter.Dispose();
				}

				log.Trace("out");
			}
		}

		private void LogException(MethodBase? metodo, Exception ex)
		{
			if (metodo == null)
				return;

			log.Error(ex.Message);
			logger.LogError(metodo.Name + "> " + ex.Message);

			if (ex.InnerException != null)
			{
				log.Error(ex.InnerException.Message);
				logger.LogError(metodo.Name + ".Inner1> " + ex.InnerException.Message);

				if (ex.InnerException.InnerException != null)
				{
					log.Error(ex.InnerException.InnerException.Message);
					logger.LogError(metodo.Name + ".Inner2> " + ex.InnerException.InnerException.Message);
				}
			}
		}
	}
}
