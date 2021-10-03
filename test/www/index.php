<html>
    <head>
        <title>PHP OwO Test</title>
        <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-EVSTQN3/azprG1Anm3QDgpJLIm9Nao0Yz1ztcQTwFspd3yD65VohhpuuCOmLASjC" crossorigin="anonymous">
    </head>
    <body>
    <div class="w-100 text-center">
        <?php
            echo '<h1 class="mb-4 mt-3">Henlo Fren OwO</h1>';
        ?>
        <img
            src='data:image/png;base64,
            <?php 
                echo base64_encode(file_get_contents("royalty-free-puppy.png"));
            ?>'
        >
        <?php
            echo "<br><br>Some random numbers to check caching<br><br>";
            echo rand();
            echo "\t\t";
            echo rand();
            echo "\t\t";
            echo rand();
        ?>
    </div>
    </body>
</html>