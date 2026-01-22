import { Injectable } from '@angular/core';
import { Meta, Title } from '@angular/platform-browser';

/**
 * Interfaz para configuración de meta tags
 */
export interface MetaTagsConfig {
  title?: string;
  description?: string;
  keywords?: string;
  author?: string;
  robots?: string;
  ogType?: string;
  ogUrl?: string;
  ogTitle?: string;
  ogDescription?: string;
  ogImage?: string;
  ogLocale?: string;
  twitterCard?: string;
  twitterTitle?: string;
  twitterDescription?: string;
  twitterImage?: string;
  canonical?: string;
}

/**
 * Servicio para gestionar meta tags dinámicos y SEO
 * Permite actualizar títulos, descripciones y Open Graph tags
 */
@Injectable({
  providedIn: 'root'
})
export class MetaTagsService {

  private readonly defaultConfig: MetaTagsConfig = {
    title: 'Soriano Mediadores - Sistema CRM',
    description: 'Sistema CRM moderno y seguro para la gestión integral de mediadores de seguros. Gestión de clientes, recobros, reportes y más.',
    keywords: 'CRM, mediadores, seguros, gestión clientes, recobros, Soriano',
    author: 'Soriano Mediadores de Seguros',
    robots: 'index, follow',
    ogType: 'website',
    ogUrl: 'https://sorianomediadores.com/',
    ogLocale: 'es_ES',
    twitterCard: 'summary_large_image'
  };

  constructor(
    private meta: Meta,
    private titleService: Title
  ) {}

  /**
   * Actualiza todos los meta tags según la configuración
   */
  updateMetaTags(config: MetaTagsConfig): void {
    const fullConfig = { ...this.defaultConfig, ...config };

    // Title
    if (fullConfig.title) {
      this.titleService.setTitle(fullConfig.title);
    }

    // Primary Meta Tags
    this.updateTag('name', 'description', fullConfig.description);
    this.updateTag('name', 'keywords', fullConfig.keywords);
    this.updateTag('name', 'author', fullConfig.author);
    this.updateTag('name', 'robots', fullConfig.robots);

    // Open Graph
    this.updateTag('property', 'og:type', fullConfig.ogType);
    this.updateTag('property', 'og:url', fullConfig.ogUrl || fullConfig.canonical);
    this.updateTag('property', 'og:title', fullConfig.ogTitle || fullConfig.title);
    this.updateTag('property', 'og:description', fullConfig.ogDescription || fullConfig.description);
    this.updateTag('property', 'og:image', fullConfig.ogImage);
    this.updateTag('property', 'og:locale', fullConfig.ogLocale);

    // Twitter
    this.updateTag('name', 'twitter:card', fullConfig.twitterCard);
    this.updateTag('name', 'twitter:title', fullConfig.twitterTitle || fullConfig.title);
    this.updateTag('name', 'twitter:description', fullConfig.twitterDescription || fullConfig.description);
    this.updateTag('name', 'twitter:image', fullConfig.twitterImage || fullConfig.ogImage);

    // Canonical URL
    if (fullConfig.canonical) {
      this.updateCanonicalUrl(fullConfig.canonical);
    }
  }

  /**
   * Actualiza o crea un meta tag específico
   */
  private updateTag(attrName: string, attrValue: string, content?: string): void {
    if (!content) {
      return;
    }

    const selector = `${attrName}="${attrValue}"`;
    const existingTag = this.meta.getTag(selector);

    if (existingTag) {
      this.meta.updateTag({ [attrName]: attrValue, content });
    } else {
      this.meta.addTag({ [attrName]: attrValue, content });
    }
  }

  /**
   * Actualiza la URL canónica
   */
  private updateCanonicalUrl(url: string): void {
    // Buscar link canonical existente
    let link: HTMLLinkElement | null = document.querySelector('link[rel="canonical"]');

    if (link) {
      link.href = url;
    } else {
      // Crear nuevo link canonical
      link = document.createElement('link');
      link.setAttribute('rel', 'canonical');
      link.setAttribute('href', url);
      document.head.appendChild(link);
    }
  }

  /**
   * Añade structured data JSON-LD a la página
   */
  addStructuredData(data: any): void {
    // Buscar script existente
    let script: HTMLScriptElement | null = document.querySelector('script[type="application/ld+json"]');

    if (!script) {
      script = document.createElement('script');
      script.type = 'application/ld+json';
      document.head.appendChild(script);
    }

    script.textContent = JSON.stringify(data);
  }

  /**
   * Crea structured data para Organization
   */
  createOrganizationStructuredData(): any {
    return {
      '@context': 'https://schema.org',
      '@type': 'Organization',
      'name': 'Soriano Mediadores de Seguros',
      'url': 'https://sorianomediadores.com',
      'logo': 'https://sorianomediadores.com/assets/logo.png',
      'description': 'Mediadores de seguros especializados en gestión integral de pólizas y recobros',
      'contactPoint': {
        '@type': 'ContactPoint',
        'telephone': '+34-XXX-XXX-XXX',
        'contactType': 'customer service',
        'areaServed': 'ES',
        'availableLanguage': ['Spanish']
      },
      'sameAs': [
        'https://www.facebook.com/sorianomediadores',
        'https://www.twitter.com/sorianomediadores',
        'https://www.linkedin.com/company/sorianomediadores'
      ]
    };
  }

  /**
   * Crea structured data para WebApplication
   */
  createWebApplicationStructuredData(): any {
    return {
      '@context': 'https://schema.org',
      '@type': 'WebApplication',
      'name': 'Soriano Mediadores CRM',
      'applicationCategory': 'BusinessApplication',
      'operatingSystem': 'Web Browser',
      'offers': {
        '@type': 'Offer',
        'price': '0',
        'priceCurrency': 'EUR'
      },
      'featureList': [
        'Gestión de clientes',
        'Control de recobros',
        'Informes y estadísticas',
        'Asistentes virtuales AI',
        'Importación de datos'
      ]
    };
  }

  /**
   * Crea structured data para BreadcrumbList
   */
  createBreadcrumbStructuredData(items: Array<{ name: string; url: string }>): any {
    return {
      '@context': 'https://schema.org',
      '@type': 'BreadcrumbList',
      'itemListElement': items.map((item, index) => ({
        '@type': 'ListItem',
        'position': index + 1,
        'name': item.name,
        'item': item.url
      }))
    };
  }
}
