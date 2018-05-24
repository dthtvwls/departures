<html>
  <head>
    <title>Departures</title>
    <style>
      body {
        color: #333;
        font: 14px "-apple-system", "Helvetica Neue", "Roboto", "Segoe UI", sans-serif;
      }
      table {
        border-collapse: collapse;
      }
      td {
        margin: -1px;
        padding: 16px;
        border-bottom: 1px #ddd solid;
      }
    </style>
  </head>
  <body>
    <?php
    function departures($station, $direction1_only = false) {
      $json = json_decode(file_get_contents("https://mtasubwaytime.info/getTime/$station"), true);
      return $direction1_only ? $json['direction1']['times'] :
        array_merge($json['direction1']['times'], $json['direction2']['times']);
    }

    $departures = array_merge(
      departures('1/142', true),
      departures('2/230'),
      departures('4/420'),
      departures('A/A38'),
      departures('E/E01'),
      departures('J/M23'),
      departures('R/R27')
    );

    usort($departures, function ($a, $b) {
      return $a['minutes'] > $b['minutes'];
    });

    echo "<table>";

    foreach ($departures as $departure) {
      echo "<tr><td><img src='http://subwaytime.mta.info/img/${departure['route']}_sm.png'></td><td>${departure['lastStation']}</td><td align='right'>${departure['minutes']}</td></tr>";
    }

    echo "</table>";
    ?>
  </body>
</html>
