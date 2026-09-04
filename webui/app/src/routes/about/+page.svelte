<script lang="ts">
  import { PageHeader } from '@hister/components';
</script>

<svelte:head>
  <title>Hister - About</title>
</svelte:head>

<div class="flex-1 overflow-y-auto px-4 py-6 md:px-12 md:py-10">
  <PageHeader color="hister-coral" class="mx-auto mb-8 max-w-3xl">About Hister</PageHeader>
  <div
    class="prose prose-neutral dark:prose-invert prose-headings:font-outfit prose-headings:font-extrabold prose-h1:text-text-brand prose-h2:text-text-brand prose-h3:text-text-brand prose-p:font-inter prose-p:text-text-brand-secondary prose-li:font-inter prose-li:text-text-brand-secondary prose-a:text-hister-indigo prose-strong:text-text-brand prose-hr:border-border-brand-muted mx-auto max-w-3xl"
  >
    <p>
      Hister is a general purpose web search engine providing automatic full-text indexing for
      visited websites.
    </p>

    <h2>Features</h2>

    <ul>
      <li>
        <strong>Privacy-focused</strong>: Keep your browsing history indexed locally on your machine
      </li>
      <li>
        <strong>Full-text indexing</strong>: Search through the actual content of web pages you've
        visited
      </li>
      <li>
        <strong>Advanced search capabilities</strong>: Utilize a powerful query language for precise
        results
      </li>
      <li>
        <strong>Smart organization</strong>: Configure blacklist and priority rules for better
        control
      </li>
      <li><strong>Efficient retrieval</strong>: Use keyword aliases to quickly find content</li>
      <li>
        <strong>Offline preview</strong>: Open clutter-free, offline preview of result web pages
      </li>
    </ul>

    <hr />
    <p>Benefits of using Hister instead of free online search engines:</p>
    <ul>
      <li>Privacy</li>
      <li>No advertisements</li>
      <li>No incorrect AI suggestion</li>
      <li>No sponsored results</li>
      <li>No manipulated result order</li>
    </ul>

    <h2>Use Cases</h2>

    <p>Hister is perfect for:</p>
    <ul>
      <li><strong>Privacy protection</strong>: Manage your data, access it without tracking</li>
      <li>
        <strong>Research</strong>: Quickly find information from articles and documentation you've
        read
      </li>
      <li>
        <strong>Reference</strong>: Locate code snippets, tutorials, or solutions you've visited
        before
      </li>
      <li>
        <strong>Knowledge management</strong>: Build a searchable archive of your web browsing
      </li>
      <li>
        <strong>Productivity</strong>: Spend less time searching for "that page I saw last week"
      </li>
    </ul>

    <h2>Technology</h2>

    <p>
      Hister is built with Go and uses the Bleve search engine for full-text search capabilities.<br
      />
      The application consists of:
    </p>

    <ul>
      <li>A local web server for search interface</li>
      <li>
        Browser extensions for <a
          href="https://chromewebstore.google.com/detail/hister/cciilamhchpmbdnniabclekddabkifhb"
          >Chrome</a
        >
        and <a href="https://addons.mozilla.org/en-US/firefox/addon/hister/">Firefox</a> to automatically
        index visited pages
      </li>
      <li>A command line utility to index/import/explore the contents of Hister</li>
    </ul>

    <h2>License</h2>

    <p>Hister is free and open-source software licensed under AGPLv3 or any later version.</p>

    <h2>Community</h2>

    <p>
      Found a bug or have a suggestion? Visit the <a
        href="https://github.com/asciimoo/hister/issues">issue tracker</a
      > on GitHub.
    </p>

    <hr />
    <ul>
      <li><a href="https://hister.org/">Website</a></li>
      <li><a href="https://github.com/asciimoo/hister">GitHub</a></li>
      <li><a href="https://codeberg.org/asciimoo/hister">Codeberg</a></li>
      <li><a href="https://discord.gg/beEyuHxRSs">Discord</a></li>
      <li>#hister on IRCNet</li>
    </ul>
  </div>
</div>
